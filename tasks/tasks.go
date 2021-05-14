package tasks

import (
	"context"
	"cowin-emailer/db"
	"cowin-emailer/response"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/imroc/req"
	"github.com/joho/godotenv"
)

const FetchSlots = "slots:fetch"

func FetchSlotsTask(districtName string, week int) *asynq.Task {
	payload := map[string]interface{}{"district": districtName, "week": week}
	return asynq.NewTask(FetchSlots, payload)
}

func HandleFetchSlotsTask(ctx context.Context, t *asynq.Task) error {
	// prepare request.
	base_url := "https://cdn-api.co-vin.in/api/v2/appointment/sessions/calendarByDistrict"
	week, err := t.Payload.GetInt("week")
	if err != nil {
		return err
	}

	loc, err := time.LoadLocation("IST")
	if err != nil {
		log.Fatal(err)
	}
	currentTime := time.Now().In(loc)
	if week == 1 {
		currentTime = currentTime.AddDate(0, 0, 7)
	} else if week == 2 {
		currentTime = currentTime.AddDate(0, 0, 14)
	} else if week == 3 {
		currentTime = currentTime.AddDate(0, 0, 21)
	}
	date := currentTime.Format("02-01-2006")
	districtName, err := t.Payload.GetString("district")
	if err != nil {
		return err
	}

	// get district id
	var districtID int
	db.DB.Select("district_id").Where("district_name = ?", districtID).First(&districtID)

	// send request and deserialize response.
	param := req.Param{
		"district_id": districtID,
		"date":        date,
	}
	resp, err := req.Get(base_url, param)
	if err != nil {
		return err
	}
	var centers []response.Center
	resp.ToJSON(&centers)

	// prepare active slots.
	var newYoungSlots []db.Slot
	var newOldSlots []db.Slot
	var activeSlots map[string]db.Slot
	var activeSlotUIDs []string
	for _, center := range centers {
		for _, session := range center.Sessions {
			if session.AvailableCapacity > 0 {
				var slot strings.Builder
				slot.WriteString(strconv.Itoa(int(center.CenterID)))
				slot.WriteString(session.SessionID)
				activeSlotUIDs = append(activeSlotUIDs, slot.String())
				activeSlots[slot.String()] = db.Slot{
					MinAgeLimit: session.MinAgeLimit,
					Date:        session.Date,
					From:        center.From,
					To:          center.To,
					Address:     center.Address,
					Name:        center.Name,
					District:    center.DistrictName,
					State:       center.StateName,
					Pincode:     center.Pincode,
					UID:         slot.String(),
				}
			}
		}
	}

	// get slots that already exist in db.
	var dbSlots []db.Slot
	db.DB.Where("uid IN ?", activeSlotUIDs).Find(&dbSlots)
	var dbSlotIDs []uint
	for _, slot := range dbSlots {
		dbSlotIDs = append(dbSlotIDs, slot.ID)
	}

	// get new slot uids.
	var newSlotUIDs []string
	var queryUIDs []string
	for _, UID := range activeSlotUIDs {
		queryUIDs = append(queryUIDs, fmt.Sprintf("(%s)", UID))
	}
	queryArg := strings.Join(queryUIDs, ",")
	db.DB.Raw("SELECT t.uid from (values ?) as t(uid) left join slot i on i.uid = t.uid where i.uid is null", queryArg).Find(&newSlotUIDs)

	// arrange new slots for 18 and 45 age groups.
	for _, slotUID := range newSlotUIDs {
		slot := activeSlots[slotUID]
		if slot.MinAgeLimit == 18 {
			newYoungSlots = append(newYoungSlots, slot)
		} else {
			newOldSlots = append(newOldSlots, slot)
		}
	}

	// get users.
	var oldUsers []db.User
	var youngUsers []db.User
	db.DB.Where("age = ? and district = ?", 18, districtName).Find(&youngUsers)
	db.DB.Where("age = ? and district = ?", 45, districtName).Find(&oldUsers)

	// prepare emails.
	var youngEmails []string
	var oldEmails []string
	for _, user := range youngUsers {
		youngEmails = append(youngEmails, user.Email)
	}
	for _, user := range oldUsers {
		oldEmails = append(oldEmails, user.Email)
	}

	// send emails, create slots and link them with users.
	if sendEmails(youngEmails, newYoungSlots) {
		db.DB.Create(&newYoungSlots)
		var newYoungUserSlots []db.UserSlot
		for _, user := range youngUsers {
			for _, slot := range newYoungSlots {
				newYoungUserSlots = append(newYoungUserSlots, db.UserSlot{
					UserID: user.ID,
					SlotID: slot.ID,
				})
			}
		}
		db.DB.Create(&newYoungUserSlots)
	}
	if sendEmails(oldEmails, newOldSlots) {
		db.DB.Create(&newOldSlots)
		var newOldUserSlots []db.UserSlot
		for _, user := range oldUsers {
			for _, slot := range newOldSlots {
				newOldUserSlots = append(newOldUserSlots, db.UserSlot{
					UserID: user.ID,
					SlotID: slot.ID,
				})
			}
		}
		db.DB.Create(&newOldUserSlots)
	}

	// get users that have not been linked with open slots that exist in the db.
	var dbUsers []db.User
	db.DB.Where("NOT id in (?)", db.DB.Table("user_slots").Select("user_id").Where("slot_id IN ?", dbSlotIDs)).Find(&dbUsers)

	// prepare emails
	var dbYoungEmails []string
	var dbOldEmails []string
	for _, user := range dbUsers {
		if user.Age == 18 {
			dbYoungEmails = append(dbYoungEmails, user.Email)
		} else {
			dbOldEmails = append(dbOldEmails, user.Email)
		}
	}

	// prepare slots
	var dbYoungSlots []db.Slot
	var dbOldSlots []db.Slot
	for _, slot := range dbSlots {
		if slot.MinAgeLimit == 18 {
			dbYoungSlots = append(dbYoungSlots, slot)
		} else {
			dbOldSlots = append(dbOldSlots, slot)
		}
	}

	// send emails and link the users with the slots
	var dbYoungUserSlots []db.UserSlot
	var dbOldUserSlots []db.UserSlot
	if sendEmails(dbYoungEmails, dbYoungSlots) {
		for _, user := range dbUsers {
			for _, slot := range dbYoungSlots {
				if user.Age == 18 {
					dbYoungUserSlots = append(dbYoungUserSlots, db.UserSlot{
						UserID: user.ID,
						SlotID: slot.ID,
					})
				}
			}
		}
		db.DB.Create(&dbYoungUserSlots)
	}
	if sendEmails(dbOldEmails, dbOldSlots) {
		for _, user := range dbUsers {
			for _, slot := range dbOldSlots {
				if user.Age == 45 {
					dbOldUserSlots = append(dbOldUserSlots, db.UserSlot{
						UserID: user.ID,
						SlotID: slot.ID,
					})
				}
			}
		}
		db.DB.Create(&dbOldUserSlots)
	}

	return nil
}

func sendEmails(emails []string, slots []db.Slot) bool {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	maiUrl := os.Getenv("MAIL_URL")
	authToken := os.Getenv("MAIL_AUTH_TOKEN")
	subject := "Some subject"
	b, err := ioutil.ReadFile("email.html")
	if err != nil {
		fmt.Print(err)
	}
	baseBody := string(b)

	var mailBody strings.Builder
	mailBody.WriteString(baseBody)
	for _, slot := range slots {
		mailBody.WriteString(slot.Address)
		mailBody.WriteString(slot.Name)
		mailBody.WriteString(slot.Date)
		mailBody.WriteString(fmt.Sprintf("%s - %s", slot.From, slot.To))
	}

	body := struct {
		FromMail string   `json:"from_mail"`
		ToMails  []string `json:"to_mails"`
		Subject  string   `json:"subject"`
		Body     string   `json:"body"`
	}{
		FromMail: "contact@ieeevit.org",
		ToMails:  emails,
		Subject:  subject,
		Body:     mailBody.String(),
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Fatalf("Error occured while preparing mail body: %s", err.Error())
	}

	auth := req.Header{
		"Authorization": authToken,
	}

	resp, err := req.Post(maiUrl, auth, req.BodyJSON(jsonBody))

	respBody := struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Error   string `json:"error"`
	}{}
	if resp.Response().StatusCode == 200 {
		resp.ToJSON(&respBody)
		if !respBody.Success {
			return true
		}
	}
	return false
}

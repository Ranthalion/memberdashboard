package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"memberserver/api/models"
	"memberserver/database"
	"memberserver/payments"
	"memberserver/resourcemanager"
	"memberserver/slack"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (a API) getTiers(w http.ResponseWriter, req *http.Request) {
	tiers := a.db.GetMemberTiers()

	w.Header().Set("Content-Type", "application/json")

	j, _ := json.Marshal(tiers)
	w.Write(j)
}

func (a API) getMembers(w http.ResponseWriter, req *http.Request) {
	members := a.db.GetMembers()

	w.Header().Set("Content-Type", "application/json")

	j, _ := json.Marshal(members)
	w.Write(j)
}

// getCurrentMember returns the logged in member details
func (a API) getCurrentUserMemberInfo(w http.ResponseWriter, req *http.Request) {
	_, user, _ := strategy.AuthenticateRequest(req)

	member, err := a.db.GetMemberByEmail(user.GetUserName())

	if err != nil {
		log.Errorf("error getting member by email: %s", err)
		http.Error(w, errors.New("error getting member by email").Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	j, _ := json.Marshal(member)
	w.Write(j)
}

func (a API) getMemberByEmail(w http.ResponseWriter, req *http.Request) {
	routeVars := mux.Vars(req)

	memberEmail := routeVars["email"]

	member, err := a.db.GetMemberByEmail(memberEmail)

	if err != nil {
		log.Errorf("error getting member by email: %s", err)
		http.Error(w, errors.New("error getting member by email").Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	j, _ := json.Marshal(member)
	w.Write(j)
}

// assignRFID to the current logged in user
func (a API) assignRFIDSelf(w http.ResponseWriter, req *http.Request) {
	var assignRFIDRequest database.AssignRFIDRequest

	err := json.NewDecoder(req.Body).Decode(&assignRFIDRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, user, _ := strategy.AuthenticateRequest(req)

	r, err := a.db.SetRFIDTag(user.GetUserName(), assignRFIDRequest.RFID)
	if err != nil {
		log.Errorf("error trying to assign rfid to member: %s", err.Error())
		http.Error(w, errors.New("unable to assign rfid").Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	j, _ := json.Marshal(r)
	w.Write(j)

	go resourcemanager.PushOne(database.Member{Email: assignRFIDRequest.Email})
}

func (a API) assignRFID(w http.ResponseWriter, req *http.Request) {
	var assignRFIDRequest database.AssignRFIDRequest

	err := json.NewDecoder(req.Body).Decode(&assignRFIDRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r, err := a.db.SetRFIDTag(assignRFIDRequest.Email, assignRFIDRequest.RFID)
	if err != nil {
		log.Errorf("error trying to assign rfid to member: %s", err.Error())
		http.Error(w, errors.New("unable to assign rfid").Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	j, _ := json.Marshal(r)
	w.Write(j)

	go resourcemanager.PushOne(database.Member{Email: assignRFIDRequest.Email})
}

func (a API) refreshPayments(w http.ResponseWriter, req *http.Request) {
	payments.GetPayments()

	a.db.ApplyMemberCredits()
	a.db.UpdateMemberTiers()

	w.Header().Set("Content-Type", "application/json")
	j, _ := json.Marshal(models.EndpointSuccess{
		Ack: true,
	})
	w.Write(j)
}

func (a API) getNonMembersOnSlack(w http.ResponseWriter, req *http.Request) {
	nonMembers := slack.FindNonMembers()
	buf := bytes.NewBufferString(strings.Join(nonMembers[:], "\n"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=nonmembersOnSlack.csv")
	w.Write(buf.Bytes())
}

func (a API) addNewMember(w http.ResponseWriter, req *http.Request) {
	var newMember database.Member

	err := json.NewDecoder(req.Body).Decode(&newMember)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// granting standard membership for the day - if their paypal email address doesn't match
	//  their membership will be revoked
	newMember.Level = uint8(database.Standard)

	err = a.db.AddMembers([]database.Member{newMember})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	j, _ := json.Marshal(models.EndpointSuccess{
		Ack: true,
	})
	w.Write(j)

	_, err = a.db.SetRFIDTag(newMember.Email, newMember.RFID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	a.db.AddUserToDefaultResources(newMember.Email)

	go resourcemanager.PushOne(newMember)
}

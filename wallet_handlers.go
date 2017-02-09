package main

import (
	"strconv"
	"net/http"
	"golang.org/x/crypto/bcrypt"
)


func CheckPassword(passwd []byte, truepasswd []byte) bool {
	err := bcrypt.CompareHashAndPassword(passwd, truepasswd)
	if err == nil {
		return true
	}
	return false
}

func CreateNewWalletHandler(w http.ResponseWriter, r *http.Request) {

	depositAmount := r.FormValue("amount")
	//refundHours := r.FormValue("time")
	return_address := r.FormValue("return_to")
	securePassword := r.FormValue("password")
	notifyMethod := r.FormValue("notify_method")
	//endEmail := r.FormValue("end_user_email")
	//endCell := r.FormValue("end_user_cell")
	//notifyEmail := r.FormValue("notify_email_address")
	notifyCell := r.FormValue("notify_text_address")
	callBackURL := r.FormValue("notify_url")

	encrytped_password := EncryptPassword([]byte(securePassword))

	newRecord := Record{ReturnWallet: return_address, Password: string(encrytped_password),
		NotifyMethod: notifyMethod, NotifyValue: notifyCell, CallBackURL: callBackURL, Amount: depositAmount, Active: false, Locked: false}

	newId := CreateNewClaim(newRecord)
	thisId := strconv.FormatInt(newId, 10)

	url := "/claim/"+thisId
	http.Redirect(w, r, url, 302)
}

func EncryptPassword(passwd []byte) []byte {
	hashedPassword, err := bcrypt.GenerateFromPassword(passwd, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hashedPassword
}

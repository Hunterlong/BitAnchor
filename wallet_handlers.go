package main

import (
	"strconv"
	"net/http"
)

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

	newRecord := Record{ReturnWallet: return_address, Password: securePassword,
		NotifyMethod: notifyMethod, NotifyValue: notifyCell, CallBackURL: callBackURL, Amount: depositAmount, Active: false, Locked: false}

	newId := CreateNewClaim(newRecord)
	thisId := strconv.FormatInt(newId, 10)

	//outputJson := map[string]interface{}{"status": "success", "id": newId}
	//output, _ := json.Marshal(outputJson)

	//SendTextMessage("18054163434", "A new wallet was created! Bitcoin Address: "+address.String())

	url := "/claim/"+thisId
	http.Redirect(w, r, url, 302)
}
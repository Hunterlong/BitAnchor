package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
	"encoding/json"
	"time"
	"fmt"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
)


var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}



func ReceiveClaimHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	thisRecord, _ := FetchClaim(id)
	var tranxId *wire.ShaHash

	password := r.FormValue("password")
	address := r.FormValue("address")
	passed := CheckPassword([]byte(thisRecord.Password),[]byte(password))

	if passed {
		fmt.Println("corrected password")
		address = ""
		tranxId, success := SendClaimToAddress(address, thisRecord)
		fmt.Println(tranxId)
		fmt.Println(success)
	} else {
		fmt.Println("wrong")
	}

	response := map[string]interface{}{"status": "success", "transaction_id": tranxId.String()}

	newJsonOutput, _ := json.Marshal(response)

	w.Write(newJsonOutput)

}



func SendClaimToAddress(address string, claim Record) (*wire.ShaHash,bool) {
	var success bool = false

	trueAddress, _ := btcutil.DecodeAddress(address,&chaincfg.MainNetParams)
	trueAmount, _ := client.GetBalanceMinConf(claim.Account,3)

	output, err := client.SendFrom(claim.Account,trueAddress,trueAmount)
	if err != nil {
		success=false
	} else {
		success =true
		MarkClaimSent(claim,address,output.String())
	}
	return output, success
}


func ClaimHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	thisRecord, _ := FetchClaim(id)
	tpl.ExecuteTemplate(w, "claim.html", thisRecord)
}



func ClaimInfoSocketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		//log.Println(err)
		return
	}

	d := 1000*time.Millisecond
	for x := range time.Tick(d) {
		thisRecord, _ := FetchClaim(id)

		response := map[string]interface{}{"id": thisRecord.Id, "amount": thisRecord.Amount, "paid": thisRecord.Paid, "active": thisRecord.Active, "transaction_id": thisRecord.TransactionId}

		newJsonOutput, _ := json.Marshal(response)
		err = conn.WriteMessage(1, newJsonOutput);

		if x.IsZero() {
		}
	}
}



func ClaimInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	thisRecord, _ := FetchClaim(id)
	response := map[string]interface{}{"id": thisRecord.Id, "amount": thisRecord.Amount, "paid": thisRecord.Paid, "active": thisRecord.Active, "transaction_id": thisRecord.TransactionId}
	newJsonOutput, _ := json.Marshal(response)
	w.Write(newJsonOutput)
}
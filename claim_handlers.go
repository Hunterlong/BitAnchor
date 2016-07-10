package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
	"encoding/json"
	"time"
)


var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
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
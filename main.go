package main

import (
	"net/http"
	"github.com/btcsuite/btcrpcclient"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/skip2/go-qrcode"
	"math/rand"
	"net/url"
	"strings"
	"html/template"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"time"
	"os"
)


type Record struct {
	Id		int
	Account		string
	Wallet		string
	OutgoingWallet	string
	ReturnWallet	string
	Amount		string
	QRcodeFile	string
	TransactionId	string
	Password	string
	CallBackURL	string
	NotifyMethod	string
	NotifyValue	string
	Active		bool
	Locked		bool
	Paid		bool
	Sent		bool
}

var client *btcrpcclient.Client
var tpl = template.Must(template.New("base").ParseGlob("*.html"))
var db *sql.DB


func startChecks() {
	for {
		time.Sleep(2 * time.Second)
		go CheckUnpaidClaims()
		go CheckUnconfirmedClaims()
	}
}


func main() {
	db, _ = sql.Open("mysql", os.Getenv("mysql_user")+":"+os.Getenv("mysql_pass")+"@tcp("+os.Getenv("mysql_host")+":3306)/"+os.Getenv("mysql_db"))

	// Connect to local bitcoin core RPC server using HTTP POST mode.
	connCfg := &btcrpcclient.ConnConfig{
		Host:         os.Getenv("bitcoin_server")+":8332",
		User:         os.Getenv("bitcoin_user"),
		Pass:         os.Getenv("bitcoin_pass"),
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, _ = btcrpcclient.New(connCfg, nil)

	// Get the current block count.
	blockCount, err := client.GetBlockCount()
	if err != nil {
		panic(err)
	}
	fmt.Println("Block count: ", blockCount)

	go startChecks()

	r := mux.NewRouter()
	r.HandleFunc("/new", CreateNewWalletHandler)
	r.HandleFunc("/claim/{id}", ClaimHandler)

	r.HandleFunc("/claim_info/{id}", ClaimInfoHandler)
	r.HandleFunc("/claim_info_ws/{id}", ClaimInfoSocketHandler)

	s := http.StripPrefix("/qrcodes/", http.FileServer(http.Dir("./qrcodes/")))
	r.PathPrefix("/qrcodes/").Handler(s)

	c := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/static/").Handler(c)

	r.Handle("/", http.FileServer(http.Dir(".")))

	http.Handle("/", r)

	message := "Server is now online!"
	SendTextMessage(os.Getenv("twilio_test"), message)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("Error: " + err.Error())
	}

}


func CheckUnpaidClaims() {
	Records, _ := FetchAllUnpaidClaims()

		for _, v := range Records {
			transactionArray, _ := client.ListTransactions(v.Account)
			for _,val := range transactionArray {
				if v.Account == val.Account {
					fmt.Println("Wallet transaction matched!")
					MarkClaimPaid(val.Account, val.Fee, val.TxID)
					if v.NotifyValue != "" {
						message := "Your Bitcoin transaction has just been paid, waiting for 3 confirmations now."
						SendTextMessage(v.NotifyValue, message)
					}
				}
			}
		}
}




func CheckUnconfirmedClaims() {
	Records, _ := FetchAllPendingConfirmsClaims()

	for _,v := range Records {
		transactionArray, _ := client.ListTransactions(v.Account)
		for _,val := range transactionArray {
			if len(transactionArray) != 0 {
					if val.Confirmations >= 3 {
						MarkClaimConfirmed(val.TxID)
						if v.NotifyValue != "" {
							message := "Your Bitcoin transaction has been confirmed!"
							SendTextMessage(v.NotifyValue, message)
						}
					}
			}
		}
	}
}



func MarkClaimPaid(account string, fee *float64, transactionId string){

	stmt, _ := db.Prepare("update records set paid=true, updated_at=NOW(), transaction_id=?, fee=? where account=?")
	res, _ := stmt.Exec(transactionId, fee, account)
	affect, _ := res.RowsAffected()
	if affect==1 {
		fmt.Println("updated record to paid!")
	}
}



func MarkClaimConfirmed(transactionId string){

	stmt, _ := db.Prepare("update records set active=true, updated_at=NOW() where transaction_id=?")
	res, _ := stmt.Exec(transactionId)
	affect, _ := res.RowsAffected()
	if affect==1 {
		fmt.Println("updated record to confirmed transaction!")
	}
}



func CreateNewClaim(newRecord Record) int64 {

	newaccountName := RandomChars(32)
	client.CreateNewAccount(newaccountName)
	address, _ := client.GetAccountAddress(newaccountName)

	randomQrName := RandomChars(32)
	qrcode_string := "bitcoin:"+address.String()+"?amount="+string(newRecord.Amount)
	qrcode_png := "qrcodes/qr-"+randomQrName+".png"
	err := qrcode.WriteFile(qrcode_string, qrcode.Medium, 256, qrcode_png)
	if err!=nil {}

	stmt, _ := db.Prepare("INSERT records SET account=?,wallet=?,return_wallet=?,amount=?,password=?,qrcode_file=?,notify_method=?,notify_value=?,callback_url=?,created_at=NOW(),updated_at=NOW()")
	res, _ := stmt.Exec(newaccountName, address.String(), newRecord.ReturnWallet, newRecord.Amount, newRecord.Password, qrcode_png, newRecord.NotifyMethod, newRecord.NotifyValue, newRecord.CallBackURL )
	id, _ := res.LastInsertId()
	return id
}


func FetchAllUnpaidClaims() ([]Record, bool) {

	rows, err := db.Query("SELECT id,account,wallet,outgoing_wallet,return_wallet,amount,password,active,locked,paid,sent,transaction_id,qrcode_file FROM records WHERE paid=?", false)
	var success bool = false
	var RecordArrays []Record

	if err == sql.ErrNoRows {
		return RecordArrays, success
	} else {

		for rows.Next() {
			var recordid int
			var account, wallet, outgoing_wallet, return_wallet, amount, password, transaction_id, qrcode_file string
			var active, locked, paid, sent bool
			err := rows.Scan(&recordid, &account, &wallet, &outgoing_wallet, &return_wallet, &amount, &password, &active, &locked, &paid, &sent, &transaction_id, &qrcode_file)
			if err != nil {
				if err == sql.ErrNoRows {
					success = false
				} else {

				}
			} else {
				record := Record{Id: recordid, Account: account, Wallet: wallet, OutgoingWallet: outgoing_wallet, ReturnWallet: return_wallet, Amount: amount, Password: password, Active: active, Locked: locked, Paid: paid, Sent: sent, TransactionId: transaction_id, QRcodeFile: qrcode_file}
				RecordArrays = append(RecordArrays, record)
				success = true
			}
		}

		return RecordArrays, success
	}

}



func FetchAllPendingConfirmsClaims() ([]Record, bool) {

	rows, _ := db.Query("SELECT id,account,wallet,outgoing_wallet,return_wallet,amount,password,active,locked,paid,sent,transaction_id,qrcode_file FROM records WHERE paid=true and active=?", false)
	var success bool = false
	var RecordArrays []Record

	for rows.Next() {
		var recordid int
		var account, wallet, outgoing_wallet, return_wallet, amount, password, transaction_id, qrcode_file string
		var active, locked, paid, sent bool
		err := rows.Scan(&recordid, &account, &wallet, &outgoing_wallet, &return_wallet, &amount, &password, &active, &locked, &paid, &sent, &transaction_id, &qrcode_file)
		if err != nil {
			if err == sql.ErrNoRows {
				success = false
			} else {

			}
		} else {
			record := Record{Id: recordid, Account: account, Wallet: wallet, OutgoingWallet: outgoing_wallet, ReturnWallet: return_wallet, Amount: amount, Password: password, Active: active, Locked: locked, Paid: paid, Sent: sent, TransactionId: transaction_id, QRcodeFile: qrcode_file}
			RecordArrays = append(RecordArrays, record)
			success = true
		}
	}

	return RecordArrays, success
}




func FetchClaim(id string) (Record, bool) {
	rows := db.QueryRow("SELECT id,account,wallet,outgoing_wallet,return_wallet,amount,password,active,locked,paid,sent,transaction_id,qrcode_file,notify_method,notify_value,callback_url FROM records WHERE id=?", id)
	var recordid int
	var account,wallet,outgoing_wallet,return_wallet,amount,password,transaction_id,qrcode_file,notify_method,notify_value,call_back_url string
	var active, locked, paid, sent bool
	var success bool = false
	err := rows.Scan(&recordid, &account,&wallet,&outgoing_wallet,&return_wallet,&amount,&password,&active,&locked,&paid,&sent,&transaction_id,&qrcode_file,&notify_method,&notify_value,&call_back_url)
	if err != nil {
		if err == sql.ErrNoRows {
			success=false
		} else {

		}
	} else {
		success=true
	}
	record := Record{Id: recordid, Account: account, Wallet: wallet, OutgoingWallet: outgoing_wallet, ReturnWallet: return_wallet, Amount: amount, Password: password, Active: active, Locked: locked, Paid: paid, Sent: sent, TransactionId: transaction_id, QRcodeFile: qrcode_file, NotifyMethod: notify_method, NotifyValue: notify_value, CallBackURL: call_back_url}

	return record, success
}


func RandomChars(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}


func MakeNewWallet(){

}

func CheckWalletTransactions(){

}


func SendTextMessage(phoneNumber string, message string) {

	fmt.Println("Sending txt to: ",phoneNumber)
	// Set initial variables
	accountSid := os.Getenv("twilio_id")
	authToken := os.Getenv("twilio_pass")
	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + accountSid + "/Messages.json"

	// Build out the data for our message
	v := url.Values{}
	v.Set("To",phoneNumber)
	v.Set("From",os.Getenv("twilio_phone"))
	v.Set("Body",message)
	rb := *strings.NewReader(v.Encode())

	// Create client
	client := &http.Client{}

	req, _ := http.NewRequest("POST", urlStr, &rb)
	req.SetBasicAuth(accountSid, authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Make request
	resp, _ := client.Do(req)
	fmt.Println(resp.StatusCode)
}
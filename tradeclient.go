package main

import (
"flag"
"fmt"
"os"
"path"
"strconv"

"github.com/quickfixgo/quickfix"
"github.com/quickfixgo/quickfix/enum"
)

func (e TradeClient) OnCreate(sessionID quickfix.SessionID) {
	return
}
func (e TradeClient) OnLogon(sessionID quickfix.SessionID) {
	fmt.Printf("Session created !! Ready to rock and roll\n")
	return
}
func (e TradeClient) OnLogout(sessionID quickfix.SessionID) {
	fmt.Printf("logged out!!!!!! It's Obama's fault !!! \n")
	return
}
func (e TradeClient) FromAdmin(msg quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	// msg.Build()
	// fmt.Printf("Incoming %s\n", &msg)
	return
}
func (e TradeClient) ToAdmin(msg quickfix.Message, sessionID quickfix.SessionID) {

	messageType,err := msg.Header.GetString(quickfix.Tag(35))
	if( err !=nil){
	}
	if(messageType == "A"){
		msg.Header.SetString(quickfix.Tag(96),"12345678")
		msg.Header.SetInt(quickfix.Tag(95),8)
		t34,err  := msg.Header.GetInt(quickfix.Tag(34))
		if(t34 == 1 && err == nil){
			msg.Header.SetBool(quickfix.Tag(141),true)  //Set reset sequence
		}
	}
	return
}

func (e TradeClient) ToApp(msg quickfix.Message, sessionID quickfix.SessionID) (err error) {
	msg.Build()
	fmt.Printf("Sending %s\n", &msg)
	return
}

//FromApp implemented as part of Application interface. This is the callback for all Application level messages from the counter party.
func (e TradeClient) FromApp(msg quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	// msg.Build()
	// fmt.Printf("Receiving %s\n", &msg)
	messageType,err := msg.Header.GetString(quickfix.Tag(35))
	if( err !=nil){
		return
	}
	switch{
	/*  EXECUTION REPORT -EXECUTION REPORT -EXECUTION REPORT -EXECUTION REPORT*/
	case messageType == string(enum.MsgType_EXECUTION_REPORT):
		execrefID ,_ := msg.Body.GetString(quickfix.Tag(20)) 
		// account  ,_ :=  msg.Body.GetString(quickfix.Tag(1))
		// ordStatus, _ := msg.Body.GetString(quickfix.Tag(39))
		// avgPx, _ := msg.Body.GetString(quickfix.Tag(6))

		if(execrefID != "3"){  //NOT order request
		// 	totalqty,_ := msg.Body.GetString(quickfix.Tag(38))
		// 	symbol,_ := msg.Body.GetString(quickfix.Tag(55))
		// 	side , _ := msg.Body.GetString(quickfix.Tag(54))
		// 	switch {
		// 	case ordStatus == string(enum.OrdStatus_NEW):
		// 		// New Order
		// 	case ordStatus == string(enum.OrdStatus_PARTIALLY_FILLED):
		// 		//Partially filled
		// 		priceFilled,_ := msg.Body.GetInt(quickfix.Tag(31))
		// 		qtyFilled,_ := msg.Body.GetInt(quickfix.Tag(32)) // qty just got filled
		// 	case ordStatus == string(enum.OrdStatus_FILLED):
		// 		//Fully filled
		// 		priceFilled,_ := msg.Body.GetInt(quickfix.Tag(31))
		// 		qtyFilled,_ := msg.Body.GetInt(quickfix.Tag(32)) // qty just got filled
		// 	case ordStatus == string(enum.OrdStatus_REJECTED):
		// 		//rejected

			// }
		}else{
			//Order status request
			account , err := msg.Body.GetString(quickfix.Tag(1))
			numPosReports,_ := msg.Body.GetString(quickfix.Tag(16728))
			for i:= range OSRs{
				if(err == nil && account == OSRs[i].account){ // GET book order not single order status request
					OSRs[i].count, _ = strconv.Atoi(numPosReports)

					var order WorkingOrder
					order.orderID ,_ = msg.Body.GetString(quickfix.Tag(37))
					order.price, _ = msg.Body.GetString(quickfix.Tag(6))
					order.quantity,_ = msg.Body.GetString(quickfix.Tag(151)) // leaves qty
					order.ordStatus , _ = msg.Body.GetString(quickfix.Tag(39)) 
					order.symbol,_ = msg.Body.GetString(quickfix.Tag(55))
					side , _ := msg.Body.GetInt(quickfix.Tag(54))
					if( side ==1){
						order.side = "buy"
					}else {
						order.side = "sell"
					}
					productType , _ := msg.Body.GetString(quickfix.Tag(167))
					if(productType == "FUT"){
						order.productMaturity,_ = msg.Body.GetString(quickfix.Tag(200))
					}
					OSRs[i].workingOrders =  append(OSRs[i].workingOrders,order)
				}

				if(OSRs[i].count  == len(OSRs[i].workingOrders)){
					// Receive all working orders
					fmt.Printf("Receve all working orders : %d for account %s \n",len(OSRs[i].workingOrders), OSRs[i].account)
					OSRs = append(OSRs[:i], OSRs[i+1:]...)
				}
			}
		}
	/* UAP - UAP - UAP - UAP - UAP - UAP - UAP */
	case messageType == "UAP":
		// fmt.Printf("Receiving UAP \n")
		var uid,_ = msg.Body.GetString(quickfix.Tag(16710))
		for i:= range UANs {
			if(uid == UANs[i].id){
				numPosReports,_ := msg.Body.GetString(quickfix.Tag(16727))
				UANs[i].count,_= strconv.Atoi(numPosReports)

				//Create a new UAP object
				var uap UAPreport
				uap.quantity,_ =  msg.Body.GetString(quickfix.Tag(32))
				q,_:=strconv.Atoi(uap.quantity)

				if(q >0){
					uap.side ="long"
				}else {
					uap.side = "short"
					uap.quantity= string(q* (-1))
				}
				uap.accountGroup,_ = msg.Header.GetString(quickfix.Tag(50)) 
				// fmt.Println(uap.accountGroup)
				uap.price,_ = msg.Body.GetString(quickfix.Tag(31))
				uap.product,_ = msg.Body.GetString(quickfix.Tag(55))
				productType , _ := msg.Body.GetString(quickfix.Tag(167))
				if(productType == "FUT"){
					uap.productMaturity,_ = msg.Body.GetString(quickfix.Tag(200))
				}
				UANs[i].reports =append(UANs[i].reports,uap)

				//UAN Complete
				if(len(UANs[i].reports) == UANs[i].count){
					// fmt.Println("Number of positions :",len(UANs[i].reports))
					for j:= 0; j< len(UANs[i].reports);j++ {
						if(UANs[i].accountGroup != UANs[i].reports[j].accountGroup){
							UANs[i].reports = append(UANs[i].reports[:j],UANs[i].reports[j+1:]...)
							j--
						}
					}
					fmt.Printf("Number of positions :%d for trader %s \n",len(UANs[i].reports),UANs[i].accountGroup)

					UANs = append(UANs[:i],UANs[i+1:]...)
				}
				break
			}
		}
	}
	return
}
var UANs []UAN
var OSRs []OrderStatusReq

func StartQuickFix() {
	flag.Parse()

	cfgFileName := path.Join("config", "tradeclient.cfg")
	if flag.NArg() > 0 {
		cfgFileName = flag.Arg(0)
	}

	cfg, err := os.Open(cfgFileName)
	if err != nil {
		fmt.Printf("Error opening %v, %v\n", cfgFileName, err)
		return
	}

	appSettings, err := quickfix.ParseSettings(cfg)
	if err != nil {
		fmt.Println("Error reading cfg,", err)
		return
	}

	app := TradeClient{}
	fileLogFactory, err := quickfix.NewFileLogFactory(appSettings)
	// screenLogFactory := quickfix.NewScreenLogFactory()

	if err != nil {
		fmt.Println("Error creating file log factory,", err)
		return
	}

	initiator, err := quickfix.NewInitiator(app, quickfix.NewMemoryStoreFactory(), appSettings, fileLogFactory)
	if err != nil {
		fmt.Printf("Unable to create Initiator: %s\n", err)
		return
	}

	initiator.Start()
	
}

type TradeClient struct {
}
type UAN struct{
	id string
	accountGroup string
	count int
	reports []UAPreport

}
type UAPreport struct{
	id string
	accountGroup string
	quantity string
	price string
	side string
	product string
	productMaturity string

}

type OrderStatusReq struct {
	account string
	count int
	workingOrders []WorkingOrder
}
type WorkingOrder struct{
	orderID string  // Used to cancel order or request order status later
	price string
	ordStatus string
	quantity string
	symbol string
	productMaturity string
	side string
}



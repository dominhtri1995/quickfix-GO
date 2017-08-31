package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/quickfixgo/quickfix"
	"github.com/quickfixgo/quickfix/enum"
	"time"
	"golang.org/x/sync/syncmap"
	"sync"
	"encoding/json"
)

func (e TradeClient) OnCreate(sessionID quickfix.SessionID) {
	return
}
func (e TradeClient) OnLogon(sessionID quickfix.SessionID) {
	fmt.Printf(" %s  Session created !! Ready to rock and roll\n", sessionID.TargetCompID)
	if connectionHealth == false {
		fmt.Println("requesting working order")
		QueryWorkingOrder(sessionID.SenderCompID)
		connectionHealth = true
	}
	companyMap.positionUpdateTimeMap.Store(sessionID.SenderCompID,0)
	return
}
func (e TradeClient) OnLogout(sessionID quickfix.SessionID) {
	connectionHealth = false
	fmt.Printf(" %s  logged out!!!!!! It's Obama's fault !!! \n", sessionID.TargetCompID)
	return
}
func (e TradeClient) FromAdmin(msg quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	//msg.Build()
	//fmt.Printf("Incoming %s\n", &msg)
	return
}
func (e TradeClient) ToAdmin(msg quickfix.Message, sessionID quickfix.SessionID) {
	//msg.Build()
	//fmt.Printf("Outgoing %s\n", &msg)

	messageType, _ := msg.Header.GetString(quickfix.Tag(35))

	if messageType == "A" {
		switch sessionID.SenderCompID {
		case "VENUSTECH":
			msg.Header.SetString(quickfix.Tag(96), "12345678")
			msg.Header.SetInt(quickfix.Tag(95), 8)
			t34, err := msg.Header.GetInt(quickfix.Tag(34))
			if t34 == 1 && err == nil {
				msg.Header.SetBool(quickfix.Tag(141), true) //Set reset sequence
			}
		case "VENUSTECH3":
			msg.Header.SetString(quickfix.Tag(96), "12345678") //set password
			msg.Header.SetInt(quickfix.Tag(95), 8)
			t34, err := msg.Header.GetInt(quickfix.Tag(34))
			if t34 == 1 && err == nil {
				msg.Header.SetBool(quickfix.Tag(141), true) //Set reset sequence
			}
		case "VENUSTECHMB":
			msg.Header.SetString(quickfix.Tag(96), "12345678") //set password
			msg.Header.SetInt(quickfix.Tag(95), 8)
			t34, err := msg.Header.GetInt(quickfix.Tag(34))
			if t34 == 1 && err == nil {
				msg.Header.SetBool(quickfix.Tag(141), true) //Set reset sequence
			}
			//add more password for different users here with more case
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
	msg.Build()
	fmt.Printf("Receiving %s\n", &msg)
	messageType, err := msg.Header.GetString(quickfix.Tag(35))
	if err != nil {
		return
	}
	switch {
	/*  EXECUTION REPORT -EXECUTION REPORT -EXECUTION REPORT -EXECUTION REPORT*/
	case messageType == string(enum.MsgType_EXECUTION_REPORT):
		execrefID, _ := msg.Body.GetString(quickfix.Tag(20))
		account, _ := msg.Body.GetString(quickfix.Tag(1))
		//ordStatus, _ := msg.Body.GetString(quickfix.Tag(39))
		execType, _ := msg.Body.GetString(quickfix.Tag(150))

		if execrefID != "3" { //NOT order Status request

			switch {
			case execType == string(enum.ExecType_NEW):
				// New Order
				clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
				if order, ok := newOrderMap.Load(clOrdID); ok && err == nil {
					order, _ := order.(*OrderConfirmation)
					order.Status = "ok"
					order.Account = account
					extractInfoExcecutionReport(order, msg) //Get all the order-related info :trade,Price plapla
					order.channel <- *order                 //Send data back to Channel
					newOrderMap.Delete(clOrdID)             // Remove Order Status request from list
				}
				//Update working order
				companyName, _ := msg.Header.GetString(quickfix.Tag(56))
				temp, _ := companyMap.CompanyWorkingOrderMap.Load(companyName)
				workingOrderList := temp.(*ConcurrentSlice)
				var wo WorkingOrder
				extractInfoERWorkingOrder(&wo, msg)
				workingOrderList.append(&wo)
				fmt.Println("added working order for ", wo.Account)

			case execType == string(enum.ExecType_REJECTED):
				//rejected
				clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
				if order, ok := newOrderMap.Load(clOrdID); ok && err == nil {
					order, _ := order.(*OrderConfirmation)
					order.Status = "rejected"
					order.Account = account
					order.Reason, _ = msg.Body.GetString(quickfix.Tag(58))
					extractInfoExcecutionReport(order, msg)
					order.channel <- *order
					newOrderMap.Delete(clOrdID)
				}
				if order, ok := cancelAndUpdateMap.Load(clOrdID); ok && err == nil {
					order, _ := order.(*OrderConfirmation)
					order.Status = "rejected"
					order.Account = account
					order.Reason, _ = msg.Body.GetString(quickfix.Tag(58))
					extractInfoExcecutionReport(order, msg)
					order.channel <- *order
					cancelAndUpdateMap.Delete(clOrdID)
				}
			case execType == string(enum.ExecType_CANCELED):
				clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
				orderID, _ := msg.Body.GetString(quickfix.Tag(37))
				if order, ok := cancelAndUpdateMap.Load(clOrdID); ok && err == nil {
					order, _ := order.(*OrderConfirmation)
					order.Status = "ok"
					order.Account = account
					extractInfoExcecutionReport(order, msg)
					order.channel <- *order
					cancelAndUpdateMap.Delete(clOrdID)
				}
				//Update working order
				companyName, _ := msg.Header.GetString(quickfix.Tag(56))
				temp, _ := companyMap.CompanyWorkingOrderMap.Load(companyName)
				workingOrderList := temp.(*ConcurrentSlice)
				for item := range workingOrderList.Iter() {
					wo, _ := item.Value.(*WorkingOrder)
					if orderID == wo.OrderID {
						workingOrderList.remove(item.Index)
					}
				}

			case execType == string(enum.ExecType_REPLACED):
				clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
				orderID, _ := msg.Body.GetString(quickfix.Tag(37))
				if order, ok := cancelAndUpdateMap.Load(clOrdID); ok && err == nil {
					order, _ := order.(*OrderConfirmation)
					order.Status = "ok"
					order.Account = account
					extractInfoExcecutionReport(order, msg)
					order.channel <- *order
					cancelAndUpdateMap.Delete(clOrdID)
				}
				//Update working order
				companyName, _ := msg.Header.GetString(quickfix.Tag(56))
				temp, _ := companyMap.CompanyWorkingOrderMap.Load(companyName)
				workingOrderList := temp.(*ConcurrentSlice)
				var newWorkingOrder WorkingOrder
				extractInfoERWorkingOrder(&newWorkingOrder, msg)
				for item := range workingOrderList.Iter() {
					wo, _ := item.Value.(*WorkingOrder)
					if orderID == wo.OrderID {
						workingOrderList.remove(item.Index)
						workingOrderList.append(&newWorkingOrder)
					}
				}
			case execType == string(enum.ExecType_RESTATED):
				orderID, _ := msg.Body.GetString(quickfix.Tag(37))
				companyName, _ := msg.Header.GetString(quickfix.Tag(56))
				temp, _ := companyMap.CompanyWorkingOrderMap.Load(companyName)
				workingOrderList := temp.(*ConcurrentSlice)
				var newWorkingOrder WorkingOrder
				extractInfoERWorkingOrder(&newWorkingOrder, msg)
				for item := range workingOrderList.Iter() {
					wo, _ := item.Value.(*WorkingOrder)
					if orderID == wo.OrderID {
						workingOrderList.remove(item.Index)
						workingOrderList.append(&newWorkingOrder)
					}
				}
			case execType == string(enum.ExecType_PARTIAL_FILL):
				//Partially filled
				var noti TTNotification

				noti.Id, _ = msg.Body.GetString(quickfix.Tag(17))
				noti.Account, _ = msg.Header.GetString(quickfix.Tag(56))
				noti.SubAccount, _ = msg.Body.GetString(quickfix.Tag(1))
				noti.Symbol, _ = msg.Body.GetString(quickfix.Tag(55))
				noti.SecurityAltID, _ = msg.Body.GetString(quickfix.Tag(10455))
				noti.PriceFilled, _ = msg.Body.GetString(quickfix.Tag(31))
				noti.QtyFilled, _ = msg.Body.GetString(quickfix.Tag(32))
				noti.OrderID, _ = msg.Body.GetString(quickfix.Tag(37))

				notifiable, err := json.Marshal(&noti)
				if err == nil {
					//Notification functions go here
					fmt.Println(string(notifiable))
				} else {
					fmt.Println("error in notification")
				}

				multiLegReportingType, _ := msg.Body.GetString(quickfix.Tag(442))
				switch multiLegReportingType {
				case "1": //Outright fill
					companyName, _ := msg.Header.GetString(quickfix.Tag(56))
					temp, _ := companyMap.CompanyWorkingOrderMap.Load(companyName)
					workingOrderList := temp.(*ConcurrentSlice)
					for item := range workingOrderList.Iter() {
						wo, _ := item.Value.(*WorkingOrder)
						if noti.OrderID == wo.OrderID {
							wo.Quantity, _ = msg.Body.GetString(quickfix.Tag(151)) //leaves qty
							wo.FilledQuantity, _ = msg.Body.GetString(quickfix.Tag(14))
							wo.AvgPx, _ = msg.Body.GetString(quickfix.Tag(6))
						}
					}
				case "2": //Single leg fill
				case "3": //entire multileg fill summary
					companyName, _ := msg.Header.GetString(quickfix.Tag(56))
					temp, _ := companyMap.CompanyWorkingOrderMap.Load(companyName)
					workingOrderList := temp.(*ConcurrentSlice)
					for item := range workingOrderList.Iter() {
						wo, _ := item.Value.(*WorkingOrder)
						if noti.OrderID == wo.OrderID {
							wo.Quantity, _ = msg.Body.GetString(quickfix.Tag(151)) //leaves qty
							wo.FilledQuantity, _ = msg.Body.GetString(quickfix.Tag(14))
							wo.AvgPx, _ = msg.Body.GetString(quickfix.Tag(6))
						}
					}
				}

			case execType == string(enum.ExecType_FILL):
				//Fully filled
				var noti TTNotification

				noti.Id, _ = msg.Body.GetString(quickfix.Tag(17))
				noti.Account, _ = msg.Body.GetString(quickfix.Tag(56))
				noti.SubAccount, _ = msg.Body.GetString(quickfix.Tag(1))
				noti.Symbol, _ = msg.Body.GetString(quickfix.Tag(55))
				noti.SecurityAltID, _ = msg.Body.GetString(quickfix.Tag(10455))
				noti.PriceFilled, _ = msg.Body.GetString(quickfix.Tag(31))
				noti.QtyFilled, _ = msg.Body.GetString(quickfix.Tag(32))
				noti.OrderID, _ = msg.Body.GetString(quickfix.Tag(37))

				notifiable, err := json.Marshal(&noti)
				if err == nil {
					//Notification functions go here
					fmt.Println(string(notifiable))
				} else {
					fmt.Println("error in notification")
				}

				//Update working order
				companyName, _ := msg.Header.GetString(quickfix.Tag(56))
				temp, _ := companyMap.CompanyWorkingOrderMap.Load(companyName)
				workingOrderList := temp.(*ConcurrentSlice)
				for item := range workingOrderList.Iter() {
					wo, _ := item.Value.(*WorkingOrder)
					if noti.OrderID == wo.OrderID {
						workingOrderList.remove(item.Index)
					}
				}
			}
		} else {
			companyName, _ := msg.Header.GetString(quickfix.Tag(56))
			if _, ok := companyMap.CompanyWorkingOrderMap.Load(companyName); !ok {
				fmt.Println("new stuff")
				var workingOrderList ConcurrentSlice
				companyMap.CompanyWorkingOrderMap.Store(companyName, &workingOrderList)
			}
			temp, _ := companyMap.CompanyWorkingOrderMap.Load(companyName)
			workingOrderList := temp.(*ConcurrentSlice)

			//Order Status request
			//numPosReports, _ := msg.Body.GetString(quickfix.Tag(16728))
			//if orderStatusRequest.Count == 0 {
			//	orderStatusRequest.Status = "ok"
			//	orderStatusRequest.channel <- *orderStatusRequest
			//	orderStatusRequestList.remove(osr.Index)
			//	return
			//}

			var order WorkingOrder
			order.OrderID, _ = msg.Body.GetString(quickfix.Tag(37))
			order.Price, _ = msg.Body.GetString(quickfix.Tag(44))
			order.Quantity, _ = msg.Body.GetString(quickfix.Tag(151)) // leaves qty
			order.FilledQuantity, _ = msg.Body.GetString(quickfix.Tag(14))
			order.OriginalQuantity, _ = msg.Body.GetString(quickfix.Tag(38))
			order.OrdStatus, _ = msg.Body.GetString(quickfix.Tag(39))
			order.Symbol, _ = msg.Body.GetString(quickfix.Tag(55))
			order.Exchange, _ = msg.Body.GetString(quickfix.Tag(207))
			order.SideNum, _ = msg.Body.GetString(quickfix.Tag(54))
			order.OrdType, _ = msg.Body.GetString(quickfix.Tag(40))
			order.TimeInForce, _ = msg.Body.GetString(quickfix.Tag(59))
			order.SecurityID, _ = msg.Body.GetString(quickfix.Tag(48))
			order.SecurityAltID, _ = msg.Body.GetString(quickfix.Tag(10455))
			order.Text, _ = msg.Body.GetString(quickfix.Tag(58))
			order.StrikePrice, _ = msg.Body.GetString(quickfix.Tag(202))
			order.StopPrice, _ = msg.Body.GetString(quickfix.Tag(99))
			order.PutOrCall, _ = msg.Body.GetString(quickfix.Tag(201))
			order.Account, _ = msg.Body.GetString(quickfix.Tag(1))
			order.AvgPx, _ = msg.Body.GetString(quickfix.Tag(6))

			if order.SideNum == "1" {
				order.Side = "Buy"
			} else {
				order.Side = "Sell"
			}
			order.ProductType, _ = msg.Body.GetString(quickfix.Tag(167))
			if order.ProductType == "FUT" || order.ProductType == "OPT" || order.ProductType == "NRG" {
				order.ProductMaturity, _ = msg.Body.GetString(quickfix.Tag(200))
			}

			if order.ProductType == "MLEG" {
				order.SecuritySubType, _ = msg.Body.GetString(quickfix.Tag(10762))
				order.NoRelatedSymUnderlyingInstrument, _ = msg.Body.GetString(quickfix.Tag(146))
				group := getUnderlyingInstrumentGroup()
				err := msg.Body.GetGroup(group)
				if err != nil {
					fmt.Println("error reading underlying group");
				} else {
					for i := 0; i < group.Len(); i++ {
						item := group.Get(i)
						var u UnderlyingInstrumentGroup
						u.UnderlyingSecurityExchange, _ = item.GetString(quickfix.Tag(308))
						u.UnderlyingSecurityType, _ = item.GetString(quickfix.Tag(310))
						u.UnderlyingSymbol, _ = item.GetString(quickfix.Tag(311))
						u.UnderlyingMaturityMonthYear, _ = item.GetString(quickfix.Tag(313))
						u.UnderlyingMaturityDay, _ = item.GetString(quickfix.Tag(314))
						u.UnderlyingContractTerm, _ = item.GetString(quickfix.Tag(18212))
						u.UnderlyingPutOrCall, _ = item.GetString(quickfix.Tag(315))
						u.UnderlyingStrikePrice, _ = item.GetString(quickfix.Tag(316))
						u.UnderlyingOptAttribute, _ = item.GetString(quickfix.Tag(317))
						u.LegSide, _ = item.GetString(quickfix.Tag(16624))
						u.LegPrice, _ = item.GetString(quickfix.Tag(10566))
						u.RatioQty, _ = item.GetString(quickfix.Tag(319))
						u.Side, _ = item.GetString(quickfix.Tag(54))
						u.UnderlyingSecurityID, _ = item.GetString(quickfix.Tag(309))
						u.UnderlyingSecurityAltID, _ = item.GetString(quickfix.Tag(10456))
						order.NoRelatedSymGroup = append(order.NoRelatedSymGroup, &u)
					}
				}
			}
			workingOrderList.append(&order)
			fmt.Printf("Recevie Restated %d \n", len(workingOrderList.items))
		}
	case messageType == string(enum.MsgType_ORDER_CANCEL_REJECT):
		clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
		if order, ok := cancelAndUpdateMap.Load(clOrdID); ok && err == nil {
			order, _ := order.(*OrderConfirmation)
			order.Status = "rejected"
			order.Reason, _ = msg.Body.GetString(quickfix.Tag(58))
			extractInfoExcecutionReport(order, msg)
			order.channel <- *order
			cancelAndUpdateMap.Delete(clOrdID)
		}
		/* UAP - UAP - UAP - UAP - UAP - UAP - UAP */
	case messageType == "UAP":
		// fmt.Printf("Receiving UAP \n")
		var uid, _ = msg.Body.GetString(quickfix.Tag(16710))
		posReqType, _ := msg.Body.GetString(quickfix.Tag(16724))

		if uan, ok := uanMap.Load(uid); ok {
			uan, _ := uan.(*UAN)
			numPosReports, _ := msg.Body.GetString(quickfix.Tag(16727))
			uan.Count, _ = strconv.Atoi(numPosReports)
			if uan.Count == 0 {
				uan.channel <- *uan // return the result to Channel
				uanMap.Delete(uid)
				return
			}

			//Create a new UAP object
			var uap UAPreport
			uap.Quantity, _ = msg.Body.GetString(quickfix.Tag(32))
			q, _ := strconv.Atoi(uap.Quantity)

			if posReqType == "1" {
				side, _ := msg.Body.GetString(quickfix.Tag(54))
				if side == "1" {
					uap.Side = "Buy"
				} else {
					uap.Side = "Sell"
				}
				uap.OrderID, _ = msg.Body.GetString(quickfix.Tag(37))
				uap.SecondaryOrderID, _ = msg.Body.GetString(quickfix.Tag(198))
			} else {
				if q > 0 {
					uap.Side = "Buy"
				} else {
					uap.Side = "Sell"
					uap.Quantity = strconv.Itoa(q * (-1))
				}
			}

			uap.PutOrCall, _ = msg.Body.GetString(quickfix.Tag(201))
			uap.AccountGroup, _ = msg.Header.GetString(quickfix.Tag(50))
			uap.Account, _ = msg.Body.GetString(quickfix.Tag(1))
			// fmt.Println(uap.AccountGroup)
			uap.Price, _ = msg.Body.GetString(quickfix.Tag(31))
			uap.SecurityID, _ = msg.Body.GetString(quickfix.Tag(48))
			uap.SecurityAltID, _ = msg.Body.GetString(quickfix.Tag(10455))
			uap.Product, _ = msg.Body.GetString(quickfix.Tag(55))
			uap.ProductType, _ = msg.Body.GetString(quickfix.Tag(167))
			uap.StrikePrice, _ = msg.Body.GetString(quickfix.Tag(202))
			if uap.ProductType == "FUT" || uap.ProductType == "OPT" || uap.ProductType == "NRG" {
				uap.ProductMaturity, _ = msg.Body.GetString(quickfix.Tag(200))
			}
			uan.Reports = append(uan.Reports, uap)

			//UAN Complete
			fmt.Println(len(uan.Reports))
			fmt.Println(uan.Count)
			if len(uan.Reports) == uan.Count {
				if posReqType == "4" {
					companyName, _ := msg.Header.GetString(quickfix.Tag(56))
					var positionList ConcurrentSlice
					for j := 0; j < len(uan.Reports); j++ {
						positionList.append(&uan.Reports[j])
					}
					companyMap.CompanyPositionMap.Store(companyName, &positionList)
				}

				fmt.Println("Done UAP")
				uan.Status = "ok"
				uan.channel <- *uan // return the result to Channel
				uanMap.Delete(uid)
			}
		}

	case messageType == "W": //Market data request
		uid, _ := msg.Body.GetString(quickfix.Tag(262))
		if md, ok := marketDataRequestMap.Load(uid); ok {
			md, _ := md.(*MarketDataReq)
			//create market data reqest response
			md.Price, err = msg.Body.GetString(quickfix.Tag(270))
			if ( err != nil) {
				md.Price = "0" //Price not available
				md.Status = "rejected"
				md.Reason = "Price not available"
			} else {
				md.Symbol, _ = msg.Body.GetString(quickfix.Tag(55))
				md.Exchange, _ = msg.Body.GetString(quickfix.Tag(207))
				productType, _ := msg.Body.GetString(quickfix.Tag(167))
				if productType == "FUT" {
					md.ProductMaturity, _ = msg.Body.GetString(quickfix.Tag(200))
				}

				md.PriceType, _ = msg.Body.GetString(quickfix.Tag(269))
				md.Status = "ok"
			}

			md.channel <- *md
			marketDataRequestMap.Delete(uid)
		}
	case messageType == "Y": //Market data reject
		uid, _ := msg.Body.GetString(quickfix.Tag(262))
		if md, ok := marketDataRequestMap.Load(uid); ok {
			md, _ := md.(*MarketDataReq)
			md.Status = "rejected"
			md.Reason, _ = msg.Body.GetString(quickfix.Tag(58))
			md.channel <- *md
			marketDataRequestMap.Delete(uid)
		}
	case messageType == "d":
		id, _ := msg.Body.GetString(quickfix.Tag(320))
		count, _ := msg.Body.GetInt(quickfix.Tag(393))
		if sdr, ok := securityDefinitionMap.Load(id); ok {
			sdr, _ := sdr.(*SecurityDefinitionReq)
			sdr.Count = count

			var security Security
			security.Symbol, _ = msg.Body.GetString(quickfix.Tag(55))
			security.Exchange, _ = msg.Body.GetString(quickfix.Tag(207))
			security.ProductType, _ = msg.Body.GetString(quickfix.Tag(167))
			security.ProductMaturity, _ = msg.Body.GetString(quickfix.Tag(200))
			security.SecurityID, _ = msg.Body.GetString(quickfix.Tag(48))
			security.SecurityAltID, _ = msg.Body.GetString(quickfix.Tag(10455))
			security.PutOrCall, _ = msg.Body.GetString(quickfix.Tag(201))
			security.StrikePrice, _ = msg.Body.GetString(quickfix.Tag(202))

			exchTickSize, _ := msg.Body.GetString(quickfix.Tag(16552))
			exchPointValue, _ := msg.Body.GetString(quickfix.Tag(16554))
			security.ExchTickSize, _ = strconv.ParseFloat(exchTickSize, 64)
			security.ExchPointValue, _ = strconv.ParseFloat(exchPointValue, 64)
			security.NumTicktblEntries, _ = msg.Body.GetInt(quickfix.Tag(16456))

			security.TickValue, security.TickSize = calculateTickValueAndSize(security.ExchTickSize, security.ExchPointValue, 0, 0, "0", "0")
			//tag16552 int,tag16554 int,tag16456 int, tag16457 int,tag16458 string
			sdr.SecurityList = append(sdr.SecurityList, security)
			fmt.Println(sdr.Count)
			fmt.Println(len(sdr.SecurityList))
			if sdr.Count == len(sdr.SecurityList) {
				fmt.Println("receive all security definition")
				sdr.Status = "ok"
				sdr.channel <- *sdr
				securityDefinitionMap.Delete(id)
			}
		}
	}
	return
}

var orderStatusRequestList ConcurrentSlice
var uanMap syncmap.Map
var newOrderMap syncmap.Map
var cancelAndUpdateMap syncmap.Map
var marketDataRequestMap syncmap.Map
var securityDefinitionMap syncmap.Map
var companyMap UserManagement
var connectionHealth bool

func StartQuickFix() {
	flag.Parse()
	connectionHealth = true
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
	//screenLogFactory := quickfix.NewScreenLogFactory()

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

// Wrapper fro Query function in console.go, using Channel to wait for all message come through.
/*
/* Use these function to avoid using Channel in main code
/*	Sepearte TT from mistro code
/*  For easy debug and maintain code with TT
 */
func TT_PAndLSOD(id string, account string, accountGroup string, sender string) (uan UAN) {

	timeTemp, ok := companyMap.positionUpdateTimeMap.Load(sender)
	timeOld, _ := timeTemp.(time.Time)
	if time.Now().Sub(timeOld).Minutes() > 60 {
		c := make(chan UAN)
		QueryPAndLSOD(id, sender, c) // Get SOD report for position upto today
		select {
		case <-c:
			companyMap.positionUpdateTimeMap.Store(sender, time.Now())
		case <-getSpecificTimeOutChan(20):
		}
	}
	temp, ok := companyMap.CompanyPositionMap.Load(sender)
	if !ok {
		uan.Status = "rejected"
		uan.Reason = "Wrong Account"
		return uan
	}
	uan.Status = "ok"
	uan.AccountGroup = accountGroup
	slice, _ := temp.(*ConcurrentSlice)
	for item := range slice.Iter() {
		uap, _ := item.Value.(*UAPreport)
		if uap.AccountGroup == accountGroup {
			uan.Reports = append(uan.Reports, *uap)
		}
	}
	uan.Count = len(uan.Reports)

	return uan
}
func TT_Fills(id string, account string, sender string) (uan UAN) {
	c := make(chan UAN)
	QueryFills(id, account, sender, c)
	select {
	case uan = <-c:

	case <-getTimeOutChan():
		var uan UAN
		uan.Status = "rejected"
		uan.Reason = "time out"
	}
	return uan
}
func TT_NewOrderSingle(id string, account string, mistroAccount string, side string, ordType string, quantity string, limitPri string, stopPri string, symbol string, exchange string, maturity string, productType string, timeInForce string, strikePrice string, putOrCall string, broker string, sender string, target string) (ordStatus OrderConfirmation) {
	c := make(chan OrderConfirmation)
	QueryNewOrderSingle(id, account, mistroAccount, side, ordType, quantity, limitPri, stopPri, symbol, exchange, maturity, productType, timeInForce, strikePrice, putOrCall, broker, sender, target, c)
	select {
	case ordStatus = <-c:
		return ordStatus
	case <-getTimeOutChan():
		ordStatus.Status = "rejected"
		ordStatus.Reason = "time out"
	}
	return ordStatus
}
func TT_NewOrderSingleAltID(id string, account string, mistroAccount string, side string, ordType string, quantity string, limitPri string, stopPri string, symbol string, exchange string, securityAltID string, productType string, timeInForce string, broker string, sender string, target string) (ordStatus OrderConfirmation) {
	c := make(chan OrderConfirmation)
	QueryNewOrderSingleAltID(id, account, mistroAccount, side, ordType, quantity, limitPri, stopPri, symbol, exchange, securityAltID, productType, timeInForce, broker, sender, target, c)
	select {
	case ordStatus = <-c:
		return ordStatus
	case <-getTimeOutChan():
		ordStatus.Status = "rejected"
		ordStatus.Reason = "time out"
	}
	return ordStatus
}

func TT_MultiLegNewOrder(id string, account string, mistroAccount, side string, ordType string, quantity string, limitPri string, stopPri string, timeInForce string, exchange string, securitySubType string, underlyingInstrumentGroup []*UnderlyingInstrumentGroup, sender string) (ordStatus OrderConfirmation) {
	c := make(chan OrderConfirmation)
	QueryMultiLegNewOrder(id, account, mistroAccount, side, ordType, quantity, limitPri, stopPri, timeInForce, exchange, securitySubType, underlyingInstrumentGroup, sender, c)
	select {
	case ordStatus = <-c:
		return ordStatus
	case <-getTimeOutChan():
		ordStatus.Status = "rejected"
		ordStatus.Reason = "time out"
	}
	return ordStatus
}

func TT_MultiLegNewOrderAltid(id string, account string, mistroAccount string, side string, ordType string, quantity string, limitPri string, stopPri string, timeInForce string, exchange string, symbol string, securityAltID string, securitySubType string, underlyingInstrumentGroup []*UnderlyingInstrumentGroup, sender string, ) (ordStatus OrderConfirmation) {
	c := make(chan OrderConfirmation)
	QueryMultiLegNewOrderAltID(id, account, mistroAccount, side, ordType, quantity, limitPri, stopPri, timeInForce, exchange, symbol, securityAltID, securitySubType, underlyingInstrumentGroup, sender, c)
	select {
	case ordStatus = <-c:
		return ordStatus
	case <-getTimeOutChan():
		ordStatus.Status = "rejected"
		ordStatus.Reason = "time out"
	}
	return ordStatus
}

func TT_WorkingOrder(account string, sender string) (wo OrderStatusReq) {
	wo.Status = "ok"
	temp, ok := companyMap.CompanyWorkingOrderMap.Load(sender)
	if !ok {
		wo.Status = "rejected"
		wo.Reason = "Wrong account"
		return wo
	}
	slice, _ := temp.(*ConcurrentSlice)
	for item := range slice.Iter() {
		t, _ := item.Value.(*WorkingOrder)
		if t.Account == account {
			wo.WorkingOrders = append(wo.WorkingOrders, *t)
		}
	}
	wo.Count = len(wo.WorkingOrders)
	return
}
func TT_OrderCancel(id string, orderID string, sender string) (ordStatus OrderConfirmation) {
	c := make(chan OrderConfirmation)
	QueryOrderCancel(id, orderID, sender, c)
	select {
	case ordStatus = <-c:
		return ordStatus
	case <-getTimeOutChan():
		ordStatus.Status = "rejected"
		ordStatus.Reason = "time out"
	}
	return ordStatus
}
func TT_OrderCancelReplace(orderID string, newid string, account string, side string, ordType string, quantity string, limitPri string, stopPri string, symbol string, exchange string, maturity string, productType string, timeInForce string, strikePrice string, putOrCall string, sender string) (ordStatus OrderConfirmation) {
	c := make(chan OrderConfirmation)
	QueryOrderCancelReplace(orderID, newid, account, side, ordType, quantity, limitPri, stopPri, symbol, exchange, maturity, productType, timeInForce, strikePrice, putOrCall, sender, c)
	select {
	case ordStatus = <-c:
		return ordStatus
	case <-getTimeOutChan():
		ordStatus.Status = "rejected"
		ordStatus.Reason = "time out"
	}
	return ordStatus
}

func TT_MultiLegOrderCancelReplace(orderID string, newid string, account string, side string, ordtype string, quantity string, limitPri string, stopPri string, timeInForce string, exchange string, securitySubType string, underlyingInstrumentGroup []*UnderlyingInstrumentGroup, sender string) (ordStatus OrderConfirmation) {
	c := make(chan OrderConfirmation)
	QueryMultilegCancelReplace(orderID, newid, account, side, ordtype, quantity, limitPri, stopPri, timeInForce, exchange, securitySubType, underlyingInstrumentGroup, sender, c)

	select {
	case ordStatus = <-c:
		return ordStatus
	case <-getTimeOutChan():
		ordStatus.Status = "rejected"
		ordStatus.Reason = "time out"
	}
	return ordStatus
}

func TT_OrderCancelRepalceAltID(orderID string, newid string, account string, mistroAccount string, side string, ordtype string, quantity string, limitPri string, stopPri string, symbol string, exchange string, securityAltID string, productType string, timeInForce string, broker string, sender string, target string)(ordStatus OrderConfirmation)  {
	c := make(chan OrderConfirmation)
	QueryCancelUpdateAltID(orderID, newid, account, mistroAccount , side , ordtype , quantity , limitPri , stopPri , symbol , exchange , securityAltID , productType , timeInForce, broker, sender, target, c)
	select {
	case ordStatus = <-c:
		return ordStatus
	case <-getTimeOutChan():
		ordStatus.Status = "rejected"
		ordStatus.Reason = "time out"
	}
	return ordStatus
}

func TT_MarketDataRequest(id string, requestType enum.SubscriptionRequestType, marketDepth int, priceType enum.MDEntryType, symbol string, exchange string, maturity string, productType string, sender string) (mdr MarketDataReq) {
	c := make(chan MarketDataReq)
	QueryMarketDataRequest(id, requestType, marketDepth, priceType, symbol, exchange, maturity, productType, sender, c)
	select {
	case mdr = <-c:
		return mdr
	case <-getTimeOutChan():
		mdr.Status = "rejected"
		mdr.Reason = "time out"
	}
	return mdr
}
func TT_QuerySecurityDefinitionRequest(id string, symbol string, exchange string, securityID string, productType string, sender string) (sdr SecurityDefinitionReq) {
	c := make(chan SecurityDefinitionReq)
	QuerySecurityDefinitionRequest(id, symbol, exchange, securityID, productType, sender, c)
	select {
	case sdr = <-c:
		return sdr
	case <-getTimeOutChan():
		sdr.Status = "rejected"
		sdr.Reason = "time out or no security found"
	}

	return sdr
}

func getTimeOutChan() chan bool {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(60 * time.Second)
		timeout <- true
	}()
	return timeout
}
func getSpecificTimeOutChan(sec time.Duration) chan bool {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(sec * time.Second)
		timeout <- true
	}()
	return timeout
}
func calculateTickValueAndSize(tag16552 float64, tag16554 float64, tag16456 int, tag16457 int, tag16458 string, price string) (tickValue float64, tickSize float64) {
	if tag16456 == 0 {
		tickSize = tag16552
		tickValue = tag16552 * tag16554
	} else {
		if price == "0" {
			return -1, -1
		}
		//baseTickSize := tag16552
		//p := float64(Price)
		//maxPirce := float64(tag16458)
		//for i := 0; i < tag16456; i++ {
		//	if p < maxPirce {
		//		TickSize = baseTickSize * tag16457
		//		TickValue = TickSize * tag16554
		//		break
		//	}
		//}
	}

	return tickValue, tickSize
}

func extractInfoExcecutionReport(order *OrderConfirmation, msg quickfix.Message) {
	order.Price, _ = msg.Body.GetString(quickfix.Tag(44)) //not available for market order
	order.StopPrice, _ = msg.Body.GetString(quickfix.Tag(99))
	order.Quantity, _ = msg.Body.GetString(quickfix.Tag(151)) // leaves qty
	order.Symbol, _ = msg.Body.GetString(quickfix.Tag(55))
	order.Exchange, _ = msg.Body.GetString(quickfix.Tag(207))
	order.ProductType, _ = msg.Body.GetString(quickfix.Tag(167))
	order.OrdType, _ = msg.Body.GetString(quickfix.Tag(40))
	order.SideNum, _ = msg.Body.GetString(quickfix.Tag(54))
	order.TimeInForce, _ = msg.Body.GetString(quickfix.Tag(59))
	order.SecurityID, _ = msg.Body.GetString(quickfix.Tag(48))
	order.SecurityAltID, _ = msg.Body.GetString(quickfix.Tag(10455))
	order.StrikePrice, _ = msg.Body.GetString(quickfix.Tag(202))
	order.PutOrCall, _ = msg.Body.GetString(quickfix.Tag(201))
	order.OrderID, _ = msg.Body.GetString(quickfix.Tag(37))
	order.Account, _ = msg.Body.GetString(quickfix.Tag(1))

	if order.SideNum == "1" {
		order.Side = "Buy"
	} else {
		order.Side = "Sell"
	}
	if order.ProductType == "FUT" || order.ProductType == "OPT" || order.ProductType == "NRG" {
		order.ProductMaturity, _ = msg.Body.GetString(quickfix.Tag(200))
	}
	if order.ProductType == "MLEG" {
		order.SecuritySubType, _ = msg.Body.GetString(quickfix.Tag(10762))
		order.NoRelatedSymUnderlyingInstrument, _ = msg.Body.GetString(quickfix.Tag(146))
		group := getUnderlyingInstrumentGroup()
		err := msg.Body.GetGroup(group)
		if err != nil {
			fmt.Println("error reading underlying group");
		} else {
			for i := 0; i < group.Len(); i++ {
				item := group.Get(i)
				var u UnderlyingInstrumentGroup
				u.Side, _ = item.GetString(quickfix.Tag(54))
				u.UnderlyingSecurityExchange, _ = item.GetString(quickfix.Tag(308))
				u.UnderlyingSecurityID, _ = item.GetString(quickfix.Tag(309))
				u.UnderlyingSecurityAltID, _ = item.GetString(quickfix.Tag(10456))
				u.UnderlyingSecurityType, _ = item.GetString(quickfix.Tag(310))
				u.UnderlyingSymbol, _ = item.GetString(quickfix.Tag(311))
				u.UnderlyingMaturityMonthYear, _ = item.GetString(quickfix.Tag(313))
				u.UnderlyingMaturityDay, _ = item.GetString(quickfix.Tag(314))
				u.UnderlyingContractTerm, _ = item.GetString(quickfix.Tag(18212))
				u.UnderlyingPutOrCall, _ = item.GetString(quickfix.Tag(315))
				u.UnderlyingStrikePrice, _ = item.GetString(quickfix.Tag(316))
				u.UnderlyingOptAttribute, _ = item.GetString(quickfix.Tag(317))
				u.LegSide, _ = item.GetString(quickfix.Tag(16624))
				u.LegPrice, _ = item.GetString(quickfix.Tag(10566))
				u.RatioQty, _ = item.GetString(quickfix.Tag(319))
				order.NoRelatedSymGroup = append(order.NoRelatedSymGroup, &u)
			}
		}
	}
}

func extractInfoERWorkingOrder(order *WorkingOrder, msg quickfix.Message) {
	order.Price, _ = msg.Body.GetString(quickfix.Tag(44)) //not available for market order
	order.StopPrice, _ = msg.Body.GetString(quickfix.Tag(99))
	order.Quantity, _ = msg.Body.GetString(quickfix.Tag(151)) // leaves qty
	order.Symbol, _ = msg.Body.GetString(quickfix.Tag(55))
	order.Exchange, _ = msg.Body.GetString(quickfix.Tag(207))
	order.ProductType, _ = msg.Body.GetString(quickfix.Tag(167))
	order.OrdType, _ = msg.Body.GetString(quickfix.Tag(40))
	order.SideNum, _ = msg.Body.GetString(quickfix.Tag(54))
	order.TimeInForce, _ = msg.Body.GetString(quickfix.Tag(59))
	order.SecurityID, _ = msg.Body.GetString(quickfix.Tag(48))
	order.SecurityAltID, _ = msg.Body.GetString(quickfix.Tag(10455))
	order.StrikePrice, _ = msg.Body.GetString(quickfix.Tag(202))
	order.PutOrCall, _ = msg.Body.GetString(quickfix.Tag(201))
	order.OrderID, _ = msg.Body.GetString(quickfix.Tag(37))
	order.Account, _ = msg.Body.GetString(quickfix.Tag(1))
	order.AvgPx, _ = msg.Body.GetString(quickfix.Tag(6))

	if order.SideNum == "1" {
		order.Side = "Buy"
	} else {
		order.Side = "Sell"
	}
	if order.ProductType == "FUT" || order.ProductType == "OPT" || order.ProductType == "NRG" {
		order.ProductMaturity, _ = msg.Body.GetString(quickfix.Tag(200))
	}
	if order.ProductType == "MLEG" {
		order.SecuritySubType, _ = msg.Body.GetString(quickfix.Tag(10762))
		order.NoRelatedSymUnderlyingInstrument, _ = msg.Body.GetString(quickfix.Tag(146))
		group := getUnderlyingInstrumentGroup()
		err := msg.Body.GetGroup(group)
		if err != nil {
			fmt.Println("error reading underlying group");
		} else {
			for i := 0; i < group.Len(); i++ {
				item := group.Get(i)
				var u UnderlyingInstrumentGroup
				u.Side, _ = item.GetString(quickfix.Tag(54))
				u.UnderlyingSecurityExchange, _ = item.GetString(quickfix.Tag(308))
				u.UnderlyingSecurityID, _ = item.GetString(quickfix.Tag(309))
				u.UnderlyingSecurityAltID, _ = item.GetString(quickfix.Tag(10456))
				u.UnderlyingSecurityType, _ = item.GetString(quickfix.Tag(310))
				u.UnderlyingSymbol, _ = item.GetString(quickfix.Tag(311))
				u.UnderlyingMaturityMonthYear, _ = item.GetString(quickfix.Tag(313))
				u.UnderlyingMaturityDay, _ = item.GetString(quickfix.Tag(314))
				u.UnderlyingContractTerm, _ = item.GetString(quickfix.Tag(18212))
				u.UnderlyingPutOrCall, _ = item.GetString(quickfix.Tag(315))
				u.UnderlyingStrikePrice, _ = item.GetString(quickfix.Tag(316))
				u.UnderlyingOptAttribute, _ = item.GetString(quickfix.Tag(317))
				u.LegSide, _ = item.GetString(quickfix.Tag(16624))
				u.LegPrice, _ = item.GetString(quickfix.Tag(10566))
				u.RatioQty, _ = item.GetString(quickfix.Tag(319))
				order.NoRelatedSymGroup = append(order.NoRelatedSymGroup, &u)
			}
		}
	}
}

type TradeClient struct {
}
type UserManagement struct {
	CompanyWorkingOrderMap syncmap.Map
	CompanyPositionMap     syncmap.Map
	positionUpdateTimeMap  syncmap.Map
}
type UAN struct {
	Id           string
	Account      string
	AccountGroup string
	Count        int
	channel      chan UAN
	Reports      []UAPreport
	Status       string
	Reason       string
}
type UAPreport struct {
	Id               string
	AccountGroup     string
	Account          string
	Quantity         string
	Price            string
	Side             string
	Product          string
	ProductType      string
	ProductMaturity  string
	Exchange         string
	SecurityID       string
	SecurityAltID    string
	PutOrCall        string //for option only
	StrikePrice      string //for option only
	OrderID          string //for fills
	SecondaryOrderID string //for fills
}

type OrderStatusReq struct {
	Account       string
	Count         int
	channel       chan OrderStatusReq
	WorkingOrders []WorkingOrder
	Status        string
	Reason        string
}
type WorkingOrder struct {
	Account          string
	OrderID          string // Used to cancel order or request order Status later
	Price            string
	StopPrice        string
	AvgPx            string
	OrdStatus        string
	Quantity         string
	FilledQuantity   string
	OriginalQuantity string
	Symbol           string
	ProductMaturity  string
	Exchange         string
	ProductType      string
	SecurityID       string
	SecurityAltID    string
	Side             string
	SideNum          string
	OrdType          string
	TimeInForce      string
	Text             string
	PutOrCall        string //for option only
	StrikePrice      string //for option only

	SecuritySubType                  string //Multileg
	NoRelatedSymUnderlyingInstrument string //Multileg
	NoRelatedSymGroup                []*UnderlyingInstrumentGroup
}

type OrderConfirmation struct {
	Id              string
	Account         string
	Status          string
	Reason          string
	OrderID         string
	Symbol          string
	ProductMaturity string
	Exchange        string
	ProductType     string
	SecurityID      string
	SecurityAltID   string
	Side            string
	Price           string
	StopPrice       string
	Quantity        string
	TimeInForce     string
	OrdType         string
	SideNum         string
	channel         chan OrderConfirmation
	PutOrCall       string //for option only
	StrikePrice     string //for option only

	SecuritySubType                  string //Multileg
	NoRelatedSymUnderlyingInstrument string //Multileg
	NoRelatedSymGroup                []*UnderlyingInstrumentGroup
}

type MarketDataReq struct {
	Id              string
	PriceType       string
	Price           string
	Symbol          string
	ProductMaturity string
	Exchange        string
	MarketDepth     string

	channel chan MarketDataReq
	Status  string
	Reason  string
}

type SecurityDefinitionReq struct {
	Id           string
	Count        int
	channel      chan SecurityDefinitionReq
	Status       string
	Reason       string
	SecurityList []Security
}
type Security struct {
	Symbol          string
	ProductMaturity string
	Exchange        string
	ProductType     string
	SecurityID      string
	SecurityAltID   string
	PutOrCall       string //for option only
	StrikePrice     string //for option only
	Currency        string

	ExchTickSize      float64
	ExchPointValue    float64
	NumTicktblEntries int

	TickSize  float64
	TickValue float64
}
type TTNotification struct {
	Id            string //unique for this fill
	Account       string
	SubAccount    string
	OrderID       string
	Symbol        string
	SecurityAltID string
	QtyFilled     string
	PriceFilled   string
}

type UnderlyingInstrumentGroup struct {
	//https://www.tradingtechnologies.com/help/fix-adapter-reference/about-fix-messages/component-block-underlying-instrument/
	UnderlyingSecurityExchange  string
	UnderlyingSecurityType      string
	UnderlyingSymbol            string
	UnderlyingSecurityID        string
	UnderlyingSecurityAltID     string
	UnderlyingMaturityMonthYear string
	UnderlyingMaturityDay       string
	UnderlyingContractTerm      string
	UnderlyingPutOrCall         string
	UnderlyingStrikePrice       string
	UnderlyingOptAttribute      string
	LegSide                     string
	LegPrice                    string
	RatioQty                    string
	Side                        string //used for response
}

func NewUnderlyingInstrumentGroup() (*UnderlyingInstrumentGroup) {
	var u UnderlyingInstrumentGroup
	u.UnderlyingSecurityExchange = ""
	u.UnderlyingSecurityType = ""
	u.UnderlyingSymbol = ""
	u.UnderlyingMaturityMonthYear = ""
	u.UnderlyingMaturityDay = ""
	u.UnderlyingContractTerm = ""
	u.UnderlyingPutOrCall = ""
	u.UnderlyingStrikePrice = ""
	u.UnderlyingOptAttribute = ""
	u.LegSide = ""
	u.LegPrice = ""
	u.RatioQty = ""
	return &u
}

//********* Concurrent SLice ************************** //

type ConcurrentSlice struct {
	sync.RWMutex
	items []interface{}
}

// Concurrent slice item
type ConcurrentSliceItem struct {
	Index int
	Value interface{}
}

// Appends an item to the concurrent slice
func (cs *ConcurrentSlice) append(item interface{}) {
	cs.Lock()
	defer cs.Unlock()
	cs.items = append(cs.items, item)
}
func (cs *ConcurrentSlice) remove(index int) {
	cs.Lock()
	defer cs.Unlock()
	cs.items = append(cs.items[:index], cs.items[index+1:]...)
}

// Iterates over the items in the concurrent slice
// Each item is sent over a Channel, so that
// we can iterate over the slice using the builin range keyword
func (cs *ConcurrentSlice) Iter() <-chan ConcurrentSliceItem {
	c := make(chan ConcurrentSliceItem)

	f := func() {
		cs.Lock()
		defer cs.Unlock()
		for index, value := range cs.items {
			c <- ConcurrentSliceItem{index, value}
		}
		close(c)
	}
	go f()

	return c
}

//********* End Concurrent Slice ************************** //

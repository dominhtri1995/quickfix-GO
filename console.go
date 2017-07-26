package main

import (
	"time"
	"strconv"

	"github.com/quickfixgo/quickfix"
	"github.com/quickfixgo/quickfix/enum"
	"github.com/quickfixgo/quickfix/field"
	"github.com/shopspring/decimal"

	fix42nos "github.com/quickfixgo/quickfix/fix42/newordersingle"
	fix42mdq "github.com/quickfixgo/quickfix/fix42/marketdatarequest"
)

func QueryPAndLSOD(id string, accountGroup string,sender string, c chan UAN) (err error) {
	//UANS
	var u UAN
	u.id = id
	u.count=0
	u.channel =c
	u.accountGroup = accountGroup

	uanMap.Store(id,&u)

	message := quickfix.NewMessage()
	queryHeader(message.Header,sender)
	message.Header.SetString(quickfix.Tag(35), "UAN")
	message.Body.SetString(quickfix.Tag(16710), id) // uniqueID
	message.Body.SetInt(quickfix.Tag(16724), 4)
	message.Body.SetString(quickfix.Tag(263), "1")
	SendMessage(message)

	return
}
func QueryPAndLPos(id string, account string,sender string, c chan UAN) (err error) {
	//UANS
	var u UAN
	u.id = id
	u.count=0
	u.channel =c
	u.account = account
	uanMap.Store(id,&u)

	message := quickfix.NewMessage()
	queryHeader(message.Header,sender)
	message.Header.SetString(quickfix.Tag(35), "UAN")
	message.Body.SetString(quickfix.Tag(16710), id) // uniqueID
	message.Body.SetInt(quickfix.Tag(16724), 0)
	message.Body.SetString(quickfix.Tag(263), "1")
	message.Body.SetString(quickfix.Tag(1), account)

	SendMessage(message)

	return
}

func QueryNewOrderSingle(id string, account string, side string, ordtype string, quantity string, pri string, symbol string, exchange string, maturity string, productType string, timeInForce string, sender string,c chan OrderConfirmation) {

	var orderQuery OrderConfirmation
	orderQuery.id = id
	orderQuery.channel =c
	newOrderMap.Store(id,&orderQuery)


	var ordType field.OrdTypeField
	ordType.FIXString = quickfix.FIXString(ordtype)
	var sideField field.SideField
	sideField.FIXString = quickfix.FIXString(side)

	order := fix42nos.New(field.NewClOrdID(id), field.NewHandlInst("1"), field.NewSymbol(symbol), sideField, field.NewTransactTime(time.Now()), ordType)

	var productTypeField field.SecurityTypeField
	productTypeField.FIXString = quickfix.FIXString(productType)
	order.Set(productTypeField)

	var timeInForceField field.TimeInForceField
	timeInForceField.FIXString = quickfix.FIXString(timeInForce)
	order.Set(timeInForceField)  //0 for day and 1 for GTC

	qty, _ := decimal.NewFromString(quantity)
	order.SetOrderQty(qty, 2)
	if(ordtype != "1") {
		price, _ := decimal.NewFromString(pri)
		order.SetPrice(price, 2)
	}
	//INStrument Block
	order.SetSecurityExchange(exchange)
	if	productType == "FUT" || productType == "OPT" || productType == "NRG"{
		order.SetMaturityMonthYear(maturity)
	}


	order.SetAccount(account)

	message := order.ToMessage()
	queryHeader(message.Header,sender)
	SendMessage(message)
}

func QueryWorkingOrder(account string,sender string, c chan OrderStatusReq) { //Order status request
	var osq OrderStatusReq
	osq.account = account
	osq.channel = c
	osq.count =0
	orderStatusRequestList.append(&osq)

	message := quickfix.NewMessage()
	queryHeader(message.Header, sender)
	message.Header.Set(field.NewMsgType("H"))
	message.Body.Set(field.NewAccount(account))

	SendMessage(message)
}

func QueryOrderCancel(id string, orderID string,sender string, c chan OrderConfirmation) {

	var cancelOrderQuery OrderConfirmation
	cancelOrderQuery.id = id
	cancelOrderQuery.channel =c
	cancelAndUpdateMap.Store(id,&cancelOrderQuery)

	message := quickfix.NewMessage()
	queryHeader(message.Header,sender)
	message.Header.Set(field.NewMsgType("F"))
	message.Body.Set(field.NewClOrdID(id))
	message.Body.Set(field.NewOrderID(orderID))

	SendMessage(message)
}

func QueryOrderCancelReplace(orderID string, newid string, account string, side string, ordtype string, quantity string, pri string, symbol string, exchange string, maturity string,productType string, timeInForce string,sender string, c chan OrderConfirmation) {

	var cancelOrderQuery OrderConfirmation
	cancelOrderQuery.id = newid
	cancelOrderQuery.channel =c
	cancelAndUpdateMap.Store(newid,&cancelOrderQuery)

	message := quickfix.NewMessage()
	queryHeader(message.Header,sender)
	message.Header.Set(field.NewMsgType("G"))
	message.Body.Set(field.NewOrderID(orderID))
	message.Body.Set(field.NewClOrdID(newid))
	message.Body.Set(field.NewSymbol(symbol))
	message.Body.Set(field.NewTransactTime(time.Now()))

	var ordType field.OrdTypeField
	ordType.FIXString = quickfix.FIXString(ordtype)
	message.Body.Set(ordType)

	var sideField field.SideField
	sideField.FIXString = quickfix.FIXString(side)
	message.Body.Set(sideField)

	var productTypeField field.SecurityTypeField
	productTypeField.FIXString = quickfix.FIXString(productType)
	message.Body.Set(productTypeField) // "FUT" for future and "OPT" for option

	var timeInForceField field.TimeInForceField
	timeInForceField.FIXString = quickfix.FIXString(timeInForce)
	message.Body.Set(timeInForceField)  //0 for day and 1 for GTC

	qty, _ := decimal.NewFromString(quantity)
	message.Body.Set(field.NewOrderQty(qty, 2))

	price, _ := decimal.NewFromString(pri)
	message.Body.Set(field.NewPrice(price, 2))

	//INStrument Block
	message.Body.Set(field.NewSecurityExchange(exchange))
	if productType == "FUT" || productType == "OPT" || productType == "NRG"{
		message.Body.Set(field.NewMaturityMonthYear(maturity))
	}

	message.Body.Set(field.NewAccount(account))
	SendMessage(message)

}
func QueryMarketDataRequest(id string, requestType enum.SubscriptionRequestType, marketDepth int, priceType enum.MDEntryType, symbol string, exchange string, maturity string,productType string,sender string, c chan MarketDataReq) {

	var md MarketDataReq
	md.id =id
	md.channel = c
	md.symbol = symbol
	md.priceType =string(priceType)
	md.productMaturity=maturity
	md.exchange =exchange
	md.marketDepth = string(marketDepth)
	md.symbol = symbol
	marketDataRequestMap.Store(id,&md)

	message := quickfix.NewMessage()
	queryHeaderPrice(message.Header,sender)
	message.Header.Set(field.NewMsgType("V"))
	message.Body.Set(field.NewMDReqID(id))
	message.Body.Set(field.NewSubscriptionRequestType(requestType))
	message.Body.Set(field.NewMarketDepth(marketDepth))
	message.Body.Set(field.NewAggregatedBook(true))   //always true
	message.Body.Set(field.NewNoMDEntryTypes(1))      //number of price types
	message.Body.Set(field.NewMDEntryType(priceType)) // price type

	////INStrument Block
	mdr := fix42mdq.FromMessage(message)

	group := fix42mdq.NewNoRelatedSymRepeatingGroup()
	newSym := group.Add()
	newSym.SetSymbol(symbol)
	newSym.SetSecurityExchange(exchange)
	var productTypeField field.SecurityTypeField
	productTypeField.FIXString = quickfix.FIXString(productType)
	newSym.Set(productTypeField)

	m, _ := strconv.Atoi(maturity)
	newSym.SetMaturityDay(m)

	mdr.SetNoRelatedSym(group)

	SendMessage(mdr.ToMessage())
}

func QuerySecurityDefinitionRequest(id string, symbol string, exchange string, securityID string, productType string,sender string, c chan SecurityDefinitionReq) {

	var sdr SecurityDefinitionReq
	sdr.id = id
	sdr.channel =c
	securityDefinitionMap.Store(id,&sdr)

	message := quickfix.NewMessage()
	queryHeader(message.Header,sender)
	message.Header.Set(field.NewMsgType("c"))
	message.Body.Set(field.NewSecurityReqID(id))
	if symbol != ""{
		message.Body.Set(field.NewSymbol(symbol))
	}

	if productType != ""{
		var productTypeField field.SecurityTypeField
		productTypeField.FIXString = quickfix.FIXString(productType)
		message.Body.Set(productTypeField) // "FUT" for future and "OPT" for option
	}
	if securityID != ""{
		message.Body.Set(field.NewSecurityID(securityID))
	}
	if exchange != ""{
		message.Body.Set(field.NewSecurityExchange(exchange))
	}
	message.Body.SetString(quickfix.Tag(17000),"Y")
	SendMessage(message)
}

func SendMessage(message quickfix.Message) {
	var m quickfix.Messagable
	m = message
	quickfix.Send(m)
}

type header interface {
	Set(f quickfix.FieldWriter) quickfix.FieldMap
}

func queryHeader(h header,sender string) {
	h.Set(field.NewSenderCompID(sender))
	h.Set(field.NewTargetCompID("TTDEV18O"))
	h.Set(field.NewBeginString("FIX.4.2"))
}

func queryHeaderPrice(h header, sender string) {
	h.Set(field.NewSenderCompID(sender))
	h.Set(field.NewTargetCompID("TTDEV18P"))
	h.Set(field.NewBeginString("FIX.4.2"))
}

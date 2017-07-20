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

func QueryPAndLSOD(id string, accountGroup string, c chan UAN) (err error) {
	//UANS
	var u UAN
	u.id = id
	u.count=0
	u.channel =c
	u.accountGroup = accountGroup

	UANs = append(UANs, u)

	message := quickfix.NewMessage()
	queryHeader(message.Header)
	message.Header.SetString(quickfix.Tag(35), "UAN")
	message.Body.SetString(quickfix.Tag(16710), id) // uniqueID
	message.Body.SetInt(quickfix.Tag(16724), 4)
	message.Body.SetString(quickfix.Tag(263), "1")
	SendMessage(message)

	return
}
func QueryPAndLPos(id string, account string, c chan UAN) (err error) {
	//UANS
	var u UAN
	u.id = id
	u.count=0
	u.channel =c
	u.account = account
	UANs = append(UANs, u)

	message := quickfix.NewMessage()
	queryHeader(message.Header)
	message.Header.SetString(quickfix.Tag(35), "UAN")
	message.Body.SetString(quickfix.Tag(16710), id) // uniqueID
	message.Body.SetInt(quickfix.Tag(16724), 0)
	message.Body.SetString(quickfix.Tag(263), "1")
	message.Body.SetString(quickfix.Tag(1), account)

	SendMessage(message)

	return
}

func QueryNewOrderSingle(id string, account string, side enum.Side, ordtype string, quantity string, pri string, symbol string, exchange string, maturity string, productType enum.SecurityType, c chan OrderConfirmation) {

	var orderQuery OrderConfirmation
	orderQuery.id = id
	orderQuery.channel =c
	NewOrders = append(NewOrders, orderQuery)

	var ordType field.OrdTypeField
	ordType.FIXString = quickfix.FIXString(ordtype)
	order := fix42nos.New(field.NewClOrdID(id), field.NewHandlInst("1"), field.NewSymbol(symbol), field.NewSide(side), field.NewTransactTime(time.Now()), ordType)

	qty, _ := decimal.NewFromString(quantity)
	order.SetOrderQty(qty, 2)
	price, _ := decimal.NewFromString(pri)
	order.SetPrice(price, 2)

	//INStrument Block
	order.SetSecurityExchange(exchange)
	order.SetSecurityType(productType) //product type FUT for future OPT for option
	if(productType =="FUT"){
		order.SetMaturityMonthYear(maturity)
	}
	order.Set(field.NewTimeInForce("0")) /// default to 0
	order.SetAccount(account)

	message := order.ToMessage()
	queryHeader(message.Header)
	SendMessage(message)
}

func QueryWorkingOrder(account string, c chan OrderStatusReq) { //Order status request
	var osq OrderStatusReq
	osq.account = account
	osq.channel = c
	osq.count =0
	OSRs = append(OSRs, osq)

	message := quickfix.NewMessage()
	queryHeader(message.Header)
	message.Header.Set(field.NewMsgType("H"))
	message.Body.Set(field.NewAccount(account))

	SendMessage(message)
}

func QueryOrderCancel(id string, orderID string, c chan OrderConfirmation) {

	var cancelOrderQuery OrderConfirmation
	cancelOrderQuery.id = id
	cancelOrderQuery.channel =c
	CancelAndUpdateOrders = append(CancelAndUpdateOrders, cancelOrderQuery)

	message := quickfix.NewMessage()
	queryHeader(message.Header)
	message.Header.Set(field.NewMsgType("F"))
	message.Body.Set(field.NewClOrdID(id))
	message.Body.Set(field.NewOrderID(orderID))

	SendMessage(message)
}

func QueryOrderCancelReplace(orderID string, newid string, account string, side enum.Side, ordType enum.OrdType, quantity string, pri string, symbol string, exchange string, maturity string,productType enum.SecurityType, c chan OrderConfirmation) {

	var cancelOrderQuery OrderConfirmation
	cancelOrderQuery.id = newid
	cancelOrderQuery.channel =c
	CancelAndUpdateOrders = append(CancelAndUpdateOrders, cancelOrderQuery)

	message := quickfix.NewMessage()
	queryHeader(message.Header)
	message.Header.Set(field.NewMsgType("G"))
	message.Body.Set(field.NewOrderID(orderID))
	message.Body.Set(field.NewClOrdID(newid))
	message.Body.Set(field.NewSymbol(symbol))
	message.Body.Set(field.NewSide(side))
	message.Body.Set(field.NewTransactTime(time.Now()))
	message.Body.Set(field.NewOrdType(ordType))

	qty, _ := decimal.NewFromString(quantity)
	message.Body.Set(field.NewOrderQty(qty, 2))

	price, _ := decimal.NewFromString(pri)
	message.Body.Set(field.NewPrice(price, 2))

	//INStrument Block
	message.Body.Set(field.NewSecurityExchange(exchange))
	message.Body.Set(field.NewSecurityType(productType)) ///default to future
	if(productType == "FUT"){
		message.Body.Set(field.NewMaturityMonthYear(maturity))
	}

	message.Body.Set(field.NewTimeInForce("0")) /// default to 0
	message.Body.Set(field.NewAccount(account))

	SendMessage(message)

}
func QueryMarketDataRequest(id string, requestType enum.SubscriptionRequestType, marketDepth int, priceType enum.MDEntryType, symbol string, exchange string, maturity string,productType enum.SecurityType, c chan MarketDataReq) {

	var md MarketDataReq
	md.id =id
	md.channel = c
	MarketDataRequests = append(MarketDataRequests,md)

	message := quickfix.NewMessage()
	queryHeaderPrice(message.Header)
	message.Header.Set(field.NewMsgType("V"))
	message.Body.Set(field.NewMDReqID(id))
	message.Body.Set(field.NewSubscriptionRequestType(requestType))
	message.Body.Set(field.NewMarketDepth(marketDepth))
	message.Body.Set(field.NewAggregatedBook(true))   //always true
	message.Body.Set(field.NewNoMDEntryTypes(1))      //number of price types
	message.Body.Set(field.NewMDEntryType(priceType)) // price type
	//message.Body.Set(field.NewNoRelatedSym(1))

	////INStrument Block
	mdr := fix42mdq.FromMessage(message)

	group := fix42mdq.NewNoRelatedSymRepeatingGroup()
	newSym := group.Add()
	newSym.SetSymbol(symbol)
	newSym.SetSecurityExchange(exchange)
	newSym.SetSecurityType(productType)
	m, _ := strconv.Atoi(maturity)
	newSym.SetMaturityDay(m)

	mdr.SetNoRelatedSym(group)

	SendMessage(mdr.ToMessage())
}

func QuerySecurityDefinitionRequest(id string, symbol string, exchange string, securityID string) {
	message := quickfix.NewMessage()
	queryHeaderPrice(message.Header)
	message.Header.Set(field.NewMsgType("c"))
	message.Body.Set(field.NewSymbol(symbol))
	message.Body.Set(field.NewSecurityType("FUT")) //Future || CHange this if want to support option
	message.Body.Set(field.NewSecurityExchange(exchange))
	message.Body.Set(field.NewSecurityID(securityID))
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

func queryHeader(h header) {
	h.Set(field.NewSenderCompID("VENUSTECH3"))
	h.Set(field.NewTargetCompID("TTDEV18O"))
	h.Set(field.NewBeginString("FIX.4.2"))
}

func queryHeaderPrice(h header) {
	h.Set(field.NewSenderCompID("VENUSTECH"))
	h.Set(field.NewTargetCompID("TTDEV18P"))
	h.Set(field.NewBeginString("FIX.4.2"))
}

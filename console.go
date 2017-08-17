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

func QueryPAndLSOD(id string, accountGroup string, sender string, c chan UAN) (err error) {
	//UANS
	var u UAN
	u.Id = id
	u.Count = 0
	u.channel = c
	u.AccountGroup = accountGroup

	uanMap.Store(id, &u)

	message := quickfix.NewMessage()
	queryHeader(message.Header, sender)
	message.Header.SetString(quickfix.Tag(35), "UAN")
	message.Body.SetString(quickfix.Tag(16710), id) // uniqueID
	message.Body.SetInt(quickfix.Tag(16724), 4)
	message.Body.SetString(quickfix.Tag(263), "0")
	SendMessage(message)

	return
}
func QueryPAndLPos(id string, account string, sender string, c chan UAN) (err error) {
	//UANS
	var u UAN
	u.Id = id
	u.Count = 0
	u.channel = c
	u.Account = account
	uanMap.Store(id, &u)

	message := quickfix.NewMessage()
	queryHeader(message.Header, sender)
	message.Header.SetString(quickfix.Tag(35), "UAN")
	message.Body.SetString(quickfix.Tag(16710), id) // uniqueID
	message.Body.SetInt(quickfix.Tag(16724), 0)
	message.Body.SetString(quickfix.Tag(263), "0")
	message.Body.SetString(quickfix.Tag(1), account)

	SendMessage(message)

	return
}
func QueryFills(id string, account string, sender string, c chan UAN) (err error) {
	//UANS
	var u UAN
	u.Id = id
	u.Count = 0
	u.channel = c
	u.Account = account
	uanMap.Store(id, &u)

	message := quickfix.NewMessage()
	queryHeader(message.Header, sender)
	message.Header.SetString(quickfix.Tag(35), "UAN")
	message.Body.SetString(quickfix.Tag(16710), id) // uniqueID
	message.Body.SetInt(quickfix.Tag(16724), 1)
	message.Body.SetString(quickfix.Tag(263), "0")
	message.Body.SetString(quickfix.Tag(1), account)

	SendMessage(message)

	return
}

func QueryNewOrderSingle(id string, account string, mistroAccount string, side string, ordtype string, quantity string, limitPri string, stopPri string, symbol string, exchange string, maturity string, productType string, timeInForce string,strikePrice string, putOrCall string, sender string,  c chan OrderConfirmation) {

	var orderQuery OrderConfirmation
	orderQuery.Id = id
	orderQuery.channel = c
	newOrderMap.Store(id, &orderQuery)

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
	order.Set(timeInForceField) //0 for day and 1 for GTC
	if timeInForce == "6"{
		order.SetExpireDate("20170810")
	}

	qty, _ := decimal.NewFromString(quantity)
	order.SetOrderQty(qty, 2)

	limitPrice, _ := decimal.NewFromString(limitPri)
	stopPrice, _ := decimal.NewFromString(stopPri)
	switch ordtype {
	case "2":
		order.SetPrice(limitPrice, 2)
	case "3":
		order.SetStopPx(stopPrice, 2)
	case "4":
		order.SetPrice(limitPrice, 2)
		order.SetStopPx(stopPrice, 2)
	case "B":
		order.SetPrice(limitPrice, 2)
	case "O":
		order.SetPrice(limitPrice, 2)
		order.SetStopPx(stopPrice, 2)
	case "Q":
		order.SetPrice(limitPrice, 2)
	case "W":
		order.SetPrice(limitPrice, 2)
		order.SetStopPx(stopPrice, 2)
	case "J":
		order.SetStopPx(stopPrice, 2)
	case "S":
		order.SetStopPx(stopPrice, 2)
	case "T":
		order.SetStopPx(stopPrice, 2)
	case "V":
		order.SetStopPx(stopPrice, 2)
	case "X":
		order.SetStopPx(stopPrice, 2)
	}

	//INStrument Block
	order.SetSecurityExchange(exchange)
	if productType == "FUT" || productType == "OPT" || productType == "NRG" {
		order.SetMaturityMonthYear(maturity)
	}

	if productType == "OPT"{
		var putOrCallField field.PutOrCallField
		putOrCallField.FIXString = quickfix.FIXString(putOrCall)
		order.Set(putOrCallField)

		strikeP ,_ := decimal.NewFromString(strikePrice)
		order.SetStrikePrice(strikeP,2)
	}

	order.SetAccount(account)
	order.SetString(quickfix.Tag(50),mistroAccount)
	order.SetString(quickfix.Tag(16142),mistroAccount)
	order.SetString(quickfix.Tag(116),"OBO") //On behalf of
	order.SetString(quickfix.Tag(11028), "Y")

	message := order.ToMessage()
	queryHeader(message.Header, sender)
	SendMessage(message)
}

func QueryMultiLegNewOrder(id string, account string, mistroAccount string, side string, ordtype string, quantity string, limitPri string, stopPri string, timeInForce string, exchange string, securitySubType string, underlyingInstrumentGroup []*UnderlyingInstrumentGroup,sender string, c chan OrderConfirmation){

	var orderQuery OrderConfirmation
	orderQuery.Id = id
	orderQuery.channel = c
	newOrderMap.Store(id, &orderQuery)

	group := getUnderlyingInstrumentGroup()
	for _,u := range underlyingInstrumentGroup{
		newSym := group.Add()
		if u.UnderlyingSymbol != ""{
			newSym.SetString(quickfix.Tag(311),u.UnderlyingSymbol)
		}
		if u.UnderlyingSecurityType != ""{
			newSym.SetString(quickfix.Tag(310),u.UnderlyingSecurityType)
		}
		if u.UnderlyingSecurityExchange != ""{
			newSym.SetString(quickfix.Tag(308),u.UnderlyingSecurityExchange)
		}
		if u.UnderlyingMaturityMonthYear != "" {
			newSym.SetString(quickfix.Tag(313),u.UnderlyingMaturityMonthYear)
		}
		if u.UnderlyingMaturityDay != ""{
			newSym.SetString(quickfix.Tag(314),u.UnderlyingMaturityDay)
		}
		if u.RatioQty != ""{
			newSym.SetString(quickfix.Tag(319),u.RatioQty)
		}
		if u.UnderlyingContractTerm != ""{
			newSym.SetString(quickfix.Tag(18212),u.UnderlyingContractTerm)
		}
		if u.UnderlyingPutOrCall != ""{
			newSym.SetString(quickfix.Tag(315),u.UnderlyingPutOrCall)
		}
		if u.UnderlyingStrikePrice != ""{
			newSym.SetString(quickfix.Tag(316),u.UnderlyingStrikePrice)
		}
		if u.UnderlyingOptAttribute != ""{
			newSym.SetString(quickfix.Tag(317),u.UnderlyingOptAttribute)
		}
		if u.LegSide != ""{
			newSym.SetString(quickfix.Tag(16624),u.LegSide)
		}
		//newSym.SetString(quickfix.Tag(10566),u.LegPrice)
		if u.UnderlyingSecurityID != ""{
			newSym.SetString(quickfix.Tag(309),u.UnderlyingSecurityID)
		}
		if u.UnderlyingSecurityAltID != ""{
			newSym.SetString(quickfix.Tag(10456),u.UnderlyingSecurityAltID)
		}
		newSym.SetString(quickfix.Tag(318),"USD")
	}

	message := quickfix.NewMessage()
	queryHeader(message.Header, sender)
	message.Body.SetString(quickfix.Tag(1),account)
	message.Body.SetString(quickfix.Tag(50),mistroAccount)
	message.Body.SetString(quickfix.Tag(16142),mistroAccount)
	message.Body.SetString(quickfix.Tag(116),"OBO") //On behalf of
	message.Body.SetString(quickfix.Tag(11028), "Y")

	message.Header.SetString(quickfix.Tag(35),"D")
	message.Body.SetString(quickfix.Tag(167),"MLEG")
	message.Body.SetString(quickfix.Tag(207),exchange)
	message.Body.SetString(quickfix.Tag(18203),"CME")

	message.Body.SetString(quickfix.Tag(11),id)
	message.Body.SetString(quickfix.Tag(10762),securitySubType)

	message.Body.SetString(quickfix.Tag(54),side)
	message.Body.SetString(quickfix.Tag(40),ordtype)
	message.Body.SetString(quickfix.Tag(59),timeInForce)
	message.Body.SetString(quickfix.Tag(55),"ES")

	qty, _ := decimal.NewFromString(quantity)
	message.Body.Set(field.NewOrderQty(qty, 2))

	limitPrice, _ := decimal.NewFromString(limitPri)
	stopPrice, _ := decimal.NewFromString(stopPri)
	switch ordtype {
	case "2":
		message.Body.Set(field.NewPrice(limitPrice,2))
	case "3":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "4":
		message.Body.Set(field.NewPrice(limitPrice,2))
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "B":
		message.Body.Set(field.NewPrice(limitPrice,2))
	case "O":
		message.Body.Set(field.NewPrice(limitPrice,2))
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "Q":
		message.Body.Set(field.NewPrice(limitPrice,2))
	case "W":
		message.Body.Set(field.NewPrice(limitPrice,2))
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "J":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "S":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "T":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "V":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "X":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	}
	message.Body.SetGroup(group)


	SendMessage(message)
}

func QueryWorkingOrder(account string, sender string, c chan OrderStatusReq) { //Order Status request
	var osq OrderStatusReq
	osq.Account = account
	osq.channel = c
	osq.Count = 0
	orderStatusRequestList.append(&osq)

	message := quickfix.NewMessage()
	queryHeader(message.Header, sender)
	message.Header.Set(field.NewMsgType("H"))
	message.Body.Set(field.NewAccount(account))

	SendMessage(message)
}

func QueryOrderCancel(id string, orderID string, sender string, c chan OrderConfirmation) {

	var cancelOrderQuery OrderConfirmation
	cancelOrderQuery.Id = id
	cancelOrderQuery.channel = c
	cancelAndUpdateMap.Store(id, &cancelOrderQuery)

	message := quickfix.NewMessage()
	queryHeader(message.Header, sender)
	message.Header.Set(field.NewMsgType("F"))
	message.Body.Set(field.NewClOrdID(id))
	message.Body.Set(field.NewOrderID(orderID))

	SendMessage(message)
}

func QueryOrderCancelReplace(orderID string, newid string, account string, side string, ordtype string, quantity string, limitPri string, stopPri string, symbol string, exchange string, maturity string, productType string, timeInForce string,strikePrice string, putOrCall string, sender string, c chan OrderConfirmation) {

	var cancelOrderQuery OrderConfirmation
	cancelOrderQuery.Id = newid
	cancelOrderQuery.channel = c
	cancelAndUpdateMap.Store(newid, &cancelOrderQuery)

	message := quickfix.NewMessage()
	queryHeader(message.Header, sender)
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
	message.Body.Set(timeInForceField) //0 for day and 1 for GTC

	qty, _ := decimal.NewFromString(quantity)
	message.Body.Set(field.NewOrderQty(qty, 2))

	limitPrice, _ := decimal.NewFromString(limitPri)
	stopPrice, _ := decimal.NewFromString(stopPri)
	switch ordtype {
	case "2":
		message.Body.Set(field.NewPrice(limitPrice,2))
	case "3":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "4":
		message.Body.Set(field.NewPrice(limitPrice,2))
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "B":
		message.Body.Set(field.NewPrice(limitPrice,2))
	case "O":
		message.Body.Set(field.NewPrice(limitPrice,2))
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "Q":
		message.Body.Set(field.NewPrice(limitPrice,2))
	case "W":
		message.Body.Set(field.NewPrice(limitPrice,2))
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "J":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "S":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "T":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "V":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "X":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	}

	//INStrument Block
	message.Body.Set(field.NewSecurityExchange(exchange))
	if productType == "FUT" || productType == "OPT" || productType == "NRG" {
		message.Body.Set(field.NewMaturityMonthYear(maturity))
	}
	if productType == "OPT"{
		var putOrCallField field.PutOrCallField
		putOrCallField.FIXString = quickfix.FIXString(putOrCall)
		message.Body.Set(putOrCallField)

		strikeP ,_ := decimal.NewFromString(strikePrice)
		message.Body.Set(field.NewStrikePrice(strikeP,2))
	}

	message.Body.Set(field.NewAccount(account))
	SendMessage(message)

}
func QueryMultilegCancelReplace(orderID string, newid string, account string, side string, ordtype string, quantity string, limitPri string, stopPri string, timeInForce string,exchange string, securitySubType string, underlyingInstrumentGroup []*UnderlyingInstrumentGroup, sender string, c chan OrderConfirmation){
	var cancelOrderQuery OrderConfirmation
	cancelOrderQuery.Id = newid
	cancelOrderQuery.channel = c
	cancelAndUpdateMap.Store(newid, &cancelOrderQuery)

	message := quickfix.NewMessage()
	queryHeader(message.Header, sender)
	message.Header.Set(field.NewMsgType("G"))
	message.Body.Set(field.NewOrderID(orderID))
	message.Body.Set(field.NewClOrdID(newid))
	message.Body.Set(field.NewTransactTime(time.Now()))

	var ordType field.OrdTypeField
	ordType.FIXString = quickfix.FIXString(ordtype)
	message.Body.Set(ordType)

	var sideField field.SideField
	sideField.FIXString = quickfix.FIXString(side)
	message.Body.Set(sideField)

	var productTypeField field.SecurityTypeField
	productTypeField.FIXString = quickfix.FIXString("MLEG")
	message.Body.Set(productTypeField) // "FUT" for future and "OPT" for option

	var timeInForceField field.TimeInForceField
	timeInForceField.FIXString = quickfix.FIXString(timeInForce)
	message.Body.Set(timeInForceField) //0 for day and 1 for GTC

	qty, _ := decimal.NewFromString(quantity)
	message.Body.Set(field.NewOrderQty(qty, 2))

	limitPrice, _ := decimal.NewFromString(limitPri)
	stopPrice, _ := decimal.NewFromString(stopPri)
	switch ordtype {
	case "2":
		message.Body.Set(field.NewPrice(limitPrice,2))
	case "3":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "4":
		message.Body.Set(field.NewPrice(limitPrice,2))
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "B":
		message.Body.Set(field.NewPrice(limitPrice,2))
	case "O":
		message.Body.Set(field.NewPrice(limitPrice,2))
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "Q":
		message.Body.Set(field.NewPrice(limitPrice,2))
	case "W":
		message.Body.Set(field.NewPrice(limitPrice,2))
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "J":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "S":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "T":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "V":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	case "X":
		message.Body.Set(field.NewStopPx(stopPrice, 2))
	}

	//INStrument Block
	message.Body.Set(field.NewSecurityExchange(exchange))
	group := getUnderlyingInstrumentGroup()
	for _,u := range underlyingInstrumentGroup{
		newSym := group.Add()
		if u.UnderlyingSymbol != ""{
			newSym.SetString(quickfix.Tag(311),u.UnderlyingSymbol)
		}
		if u.UnderlyingSecurityType != ""{
			newSym.SetString(quickfix.Tag(310),u.UnderlyingSecurityType)
		}
		if u.UnderlyingSecurityExchange != ""{
			newSym.SetString(quickfix.Tag(308),u.UnderlyingSecurityExchange)
		}
		if u.UnderlyingMaturityMonthYear != "" {
			newSym.SetString(quickfix.Tag(313),u.UnderlyingMaturityMonthYear)
		}
		if u.UnderlyingMaturityDay != ""{
			newSym.SetString(quickfix.Tag(314),u.UnderlyingMaturityDay)
		}
		if u.RatioQty != ""{
			newSym.SetString(quickfix.Tag(319),u.RatioQty)
		}
		if u.UnderlyingContractTerm != ""{
			newSym.SetString(quickfix.Tag(18212),u.UnderlyingContractTerm)
		}
		if u.UnderlyingPutOrCall != ""{
			newSym.SetString(quickfix.Tag(315),u.UnderlyingPutOrCall)
		}
		if u.UnderlyingStrikePrice != ""{
			newSym.SetString(quickfix.Tag(316),u.UnderlyingStrikePrice)
		}
		if u.UnderlyingOptAttribute != ""{
			newSym.SetString(quickfix.Tag(317),u.UnderlyingOptAttribute)
		}
		if u.LegSide != ""{
			newSym.SetString(quickfix.Tag(16624),u.LegSide)
		}
		//newSym.SetString(quickfix.Tag(10566),u.LegPrice)
		if u.UnderlyingSecurityID != ""{
			newSym.SetString(quickfix.Tag(309),u.UnderlyingSecurityID)
		}
		if u.UnderlyingSecurityAltID != ""{
			newSym.SetString(quickfix.Tag(10456),u.UnderlyingSecurityAltID)
		}
		newSym.SetString(quickfix.Tag(318),"USD")
	}

	message.Body.SetString(quickfix.Tag(10762),securitySubType)
	message.Body.Set(field.NewAccount(account))
	message.Body.SetGroup(group)
	SendMessage(message)
}

func QueryMarketDataRequest(id string, requestType enum.SubscriptionRequestType, marketDepth int, priceType enum.MDEntryType, symbol string, exchange string, maturity string, productType string, sender string, c chan MarketDataReq) {

	var md MarketDataReq
	md.Id = id
	md.channel = c
	md.Symbol = symbol
	md.PriceType = string(priceType)
	md.ProductMaturity = maturity
	md.Exchange = exchange
	md.MarketDepth = string(marketDepth)
	md.Symbol = symbol
	marketDataRequestMap.Store(id, &md)

	message := quickfix.NewMessage()
	queryHeaderPrice(message.Header, sender)
	message.Header.Set(field.NewMsgType("V"))
	message.Body.Set(field.NewMDReqID(id))
	message.Body.Set(field.NewSubscriptionRequestType(requestType))
	message.Body.Set(field.NewMarketDepth(marketDepth))
	message.Body.Set(field.NewAggregatedBook(true))   //always true
	message.Body.Set(field.NewNoMDEntryTypes(1))      //number of Price types
	message.Body.Set(field.NewMDEntryType(priceType)) // Price type

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

func QuerySecurityDefinitionRequest(id string, symbol string, exchange string, securityID string, productType string, sender string, c chan SecurityDefinitionReq) {

	var sdr SecurityDefinitionReq
	sdr.Id = id
	sdr.channel = c
	securityDefinitionMap.Store(id, &sdr)

	message := quickfix.NewMessage()
	queryHeader(message.Header, sender)
	message.Header.Set(field.NewMsgType("c"))
	message.Body.Set(field.NewSecurityReqID(id))
	if symbol != "" {
		message.Body.Set(field.NewSymbol(symbol))
	}

	if productType != "" {
		var productTypeField field.SecurityTypeField
		productTypeField.FIXString = quickfix.FIXString(productType)
		message.Body.Set(productTypeField) // "FUT" for future and "OPT" for option
	}
	if securityID != "" {
		message.Body.Set(field.NewSecurityID(securityID))
	}
	if exchange != "" {
		message.Body.Set(field.NewSecurityExchange(exchange))
	}
	//message.Body.SetString(quickfix.Tag(17000),"Y")
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

func queryHeader(h header, sender string) {
	h.Set(field.NewSenderCompID(sender))
	h.Set(field.NewTargetCompID("TTDEV18O"))
	h.Set(field.NewBeginString("FIX.4.2"))
}

func queryHeaderPrice(h header, sender string) {
	h.Set(field.NewSenderCompID(sender))
	h.Set(field.NewTargetCompID("TTDEV18P"))
	h.Set(field.NewBeginString("FIX.4.2"))
}

func getUnderlyingInstrumentGroup() *quickfix.RepeatingGroup{
	var groupTemplate []quickfix.GroupItem
	groupTemplate= append(groupTemplate, quickfix.GroupElement(quickfix.Tag(311)))
	groupTemplate= append(groupTemplate, quickfix.GroupElement(quickfix.Tag(309)))
	groupTemplate= append(groupTemplate, quickfix.GroupElement(quickfix.Tag(310)))
	groupTemplate= append(groupTemplate, quickfix.GroupElement(quickfix.Tag(308)))
	groupTemplate= append(groupTemplate, quickfix.GroupElement(quickfix.Tag(10456)))
	groupTemplate= append(groupTemplate, quickfix.GroupElement(quickfix.Tag(318)))
	groupTemplate= append(groupTemplate, quickfix.GroupElement(quickfix.Tag(313)))
	groupTemplate= append(groupTemplate, quickfix.GroupElement(quickfix.Tag(314)))
	groupTemplate= append(groupTemplate, quickfix.GroupElement(quickfix.Tag(18212)))
	groupTemplate= append(groupTemplate, quickfix.GroupElement(quickfix.Tag(315)))
	groupTemplate= append(groupTemplate, quickfix.GroupElement(quickfix.Tag(316)))
	groupTemplate= append(groupTemplate, quickfix.GroupElement(quickfix.Tag(317)))
	groupTemplate= append(groupTemplate, quickfix.GroupElement(quickfix.Tag(319)))
	groupTemplate= append(groupTemplate, quickfix.GroupElement(quickfix.Tag(16624)))
	groupTemplate= append(groupTemplate, quickfix.GroupElement(quickfix.Tag(10566)))

	group := quickfix.NewRepeatingGroup(quickfix.Tag(146),groupTemplate)
	return group
}

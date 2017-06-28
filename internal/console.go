package internal

import (
"bufio"
"fmt"
"time"

"github.com/quickfixgo/quickfix"
"github.com/quickfixgo/quickfix/enum"
"github.com/quickfixgo/quickfix/field"
"github.com/shopspring/decimal"

"os"
"strconv"
"strings"

fix42nos "github.com/quickfixgo/quickfix/fix42/newordersingle"
fix42cxl "github.com/quickfixgo/quickfix/fix42/ordercancelrequest"
fix42mdr "github.com/quickfixgo/quickfix/fix42/marketdatarequest"
)

func GetPAndL() (err error){
	message := quickfix.NewMessage()
	message.Header.Set(field.NewSenderCompID("VENUSTECH"))
	message.Header.Set(field.NewTargetCompID("TTDEV18O"))
	message.Header.SetString(quickfix.Tag(8),"FIX.4.2")
	message.Header.SetString(quickfix.Tag(35),"UAN")
	message.Body.SetString(quickfix.Tag(16710),"14")
	message.Body.SetInt(quickfix.Tag(16724),4)
	message.Body.SetString(quickfix.Tag(263),"1")
	var m quickfix.Messagable
	m= message
	quickfix.Send(m)

	return 
}
func queryString(fieldName string) string {
	fmt.Printf("%v: ", fieldName)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return scanner.Text()
}

func queryDecimal(fieldName string) decimal.Decimal {
	val, err := decimal.NewFromString(queryString(fieldName))
	if err != nil {
		panic(err)
	}

	return val
}

func queryFieldChoices(fieldName string, choices []string, values []string) string {
	for i, choice := range choices {
		fmt.Printf("%v) %v\n", i+1, choice)
	}

	choiceStr := queryString(fieldName)
	choice, err := strconv.Atoi(choiceStr)
	if err != nil || choice < 1 || choice > len(choices) {
		panic(fmt.Errorf("Invalid %v: %v", fieldName, choice))
	}

	if values == nil {
		return choiceStr
	}

	return values[choice-1]
}

func QueryAction() (string, error) {
	fmt.Println()
	fmt.Println("1) Enter Order")
	fmt.Println("2) Cancel Order")
	fmt.Println("3) Request Market Test")
	fmt.Println("4) Quit")
	fmt.Print("Action: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}

func queryClOrdID() field.ClOrdIDField {
	return field.NewClOrdID(queryString("ClOrdID"))
}

func queryOrigClOrdID() field.OrigClOrdIDField {
	return field.NewOrigClOrdID(("OrigClOrdID"))
}

func querySymbol() field.SymbolField {
	return field.NewSymbol(queryString("Symbol"))
}

func querySide() field.SideField {
	choices := []string{
		"Buy",
		"Sell",
		"Sell Short",
		"Sell Short Exempt",
		"Cross",
		"Cross Short",
		"Cross Short Exempt",
	}

	values := []string{
		string(enum.Side_BUY),
		string(enum.Side_SELL),
		string(enum.Side_SELL_SHORT),
		string(enum.Side_SELL_SHORT_EXEMPT),
		string(enum.Side_CROSS),
		string(enum.Side_CROSS_SHORT),
		"A",
	}

	return field.NewSide(enum.Side(queryFieldChoices("Side", choices, values)))
}

func queryOrdType(f *field.OrdTypeField) field.OrdTypeField {
	choices := []string{
		"Market",
		"Limit",
		"Stop",
		"Stop Limit",
	}

	values := []string{
		string(enum.OrdType_MARKET),
		string(enum.OrdType_LIMIT),
		string(enum.OrdType_STOP),
		string(enum.OrdType_STOP_LIMIT),
	}

	f.FIXString = quickfix.FIXString(queryFieldChoices("OrdType", choices, values))
	return *f
}

func queryTimeInForce() field.TimeInForceField {
	choices := []string{
		"Day",
		"IOC",
		"OPG",
		"GTC",
		"GTX",
	}
	values := []string{
		string(enum.TimeInForce_DAY),
		string(enum.TimeInForce_IMMEDIATE_OR_CANCEL),
		string(enum.TimeInForce_AT_THE_OPENING),
		string(enum.TimeInForce_GOOD_TILL_CANCEL),
		string(enum.TimeInForce_GOOD_TILL_CROSSING),
	}

	return field.NewTimeInForce(enum.TimeInForce(queryFieldChoices("TimeInForce", choices, values)))
}

func queryOrderQty() field.OrderQtyField {
	return field.NewOrderQty(queryDecimal("OrderQty"), 2)
}

func queryPrice() field.PriceField {
	return field.NewPrice(queryDecimal("Price"), 2)
}

func queryStopPx() field.StopPxField {
	return field.NewStopPx(queryDecimal("Stop Price"), 2)
}

func querySenderCompID() field.SenderCompIDField {
	return field.NewSenderCompID(queryString("SenderCompID"))
}

func queryTargetCompID() field.TargetCompIDField {
	return field.NewTargetCompID(queryString("TargetCompID"))
}

func queryTargetSubID() field.TargetSubIDField {
	return field.NewTargetSubID(queryString("TargetSubID"))
}

func queryConfirm(prompt string) bool {
	fmt.Println()
	fmt.Printf("%v?: ", prompt)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	return strings.ToUpper(scanner.Text()) == "Y"
}

type header interface {
	Set(f quickfix.FieldWriter) quickfix.FieldMap
}

func queryHeader(h header) {
	h.Set(field.NewSenderCompID("VENUSTECH"))
	h.Set(field.NewTargetCompID("TTDEV18O"))
	h.Set(field.NewBeginString("FIX.4.2"))
}

func queryNewOrderSingle42() (msg quickfix.Message) {
	var ordType field.OrdTypeField
	ordType.FIXString = quickfix.FIXString("2")
	order := fix42nos.New(queryClOrdID(), field.NewHandlInst("1"), field.NewSymbol("BZ"), field.NewSide("1"), field.NewTransactTime(time.Now()), ordType)
	
	qty, _ := decimal.NewFromString("500")
	order.SetOrderQty(qty,2)

	price, _ := decimal.NewFromString("4615")
	order.SetPrice(price,2)

	//INStrument Block
	order.SetSecurityExchange("CME")
	order.SetSecurityID("00A0HR00BZZ")

	order.Set(field.NewTimeInForce("0"))
	order.SetAccount("venustech")
	msg = order.ToMessage()
	queryHeader(msg.Header)

	return
}

func queryOrderCancelRequest42() (msg quickfix.Message) {
	cancel := fix42cxl.New(queryOrigClOrdID(), queryClOrdID(), querySymbol(), querySide(), field.NewTransactTime(time.Now()))
	cancel.Set(queryOrderQty())
	msg = cancel.ToMessage()
	queryHeader(msg.Header)
	return
}

func queryMarketDataRequest42() fix42mdr.MarketDataRequest {
	request := fix42mdr.New(field.NewMDReqID("MARKETDATAID"),
		field.NewSubscriptionRequestType(enum.SubscriptionRequestType_SNAPSHOT),
		field.NewMarketDepth(0),
		)

	entryTypes := fix42mdr.NewNoMDEntryTypesRepeatingGroup()
	entryTypes.Add().SetMDEntryType(enum.MDEntryType_BID)
	request.SetNoMDEntryTypes(entryTypes)

	relatedSym := fix42mdr.NewNoRelatedSymRepeatingGroup()
	relatedSym.Add().SetSymbol("LNUX")
	request.SetNoRelatedSym(relatedSym)

	queryHeader(request.Header)
	return request
}

func QueryEnterOrder() (err error) {
	
	// var m quickfix.Messagable
	// m = queryNewOrderSingle42()
	// quickfix.Send(m)
	GetPAndL()
	return 
}

func QueryCancelOrder() (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
		}()

		var cxl quickfix.Message
		cxl = queryOrderCancelRequest42()


		if queryConfirm("Send Cancel") {
			return quickfix.Send(cxl)
		}

		return
}

func QueryMarketDataRequest() error {

	var req quickfix.Messagable
	req = queryMarketDataRequest42()



	if queryConfirm("Send MarketDataRequest") {
		return quickfix.Send(req)
	}

	return nil
}

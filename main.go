package main  
import (
"fmt"
"os"
"bufio"
"github.com/rs/xid"

)
func main() {
	
	StartQuickFix()
Loop:
	for {
		action, err := QueryAction()
		if err != nil {
			break
		}
		switch action {
		case "1":
			c := make(chan UAN)
			QueryPAndLSOD(xid.New().String(),"TTORDFA222222",c)

			uan := <-c
			fmt.Printf("wth")
			for _, uap := range uan.reports{
				fmt.Printf("%s",uap.product)
			}
		case "2":
			c := make (chan OrderConfirmation)
			QueryNewOrderSingle(xid.New().String(),"venustech","1","2","500","4570","BZ","CME","201709","FUT",c)
			ordStatus := <-c
			fmt.Printf("Status: %s",ordStatus.status)
		case "3":
			c := make (chan OrderStatusReq)
			QueryWorkingOrder("venustech",c)
			wo := <- c
			for _, order := range wo.workingOrders{
				fmt.Printf("%s",order.symbol)
			}
		case "4":
			c := make (chan MarketDataReq)
			QueryMarketDataRequest(xid.New().String(),"0",0,"2","BZ","CME","201709","FUT",c)
			mdr := <- c
			fmt.Printf("Price :%s",mdr.price)

		case "5":
			c := make (chan OrderStatusReq)
			QueryWorkingOrder("venustech",c) // Get all working order
			wo := <- c

			c1 := make (chan OrderConfirmation)
			QueryOrderCancel(xid.New().String(), wo.workingOrders[0].orderID, c1) // Cancel the first working order
			ordStatus := <-c1
			fmt.Printf("Status: %s",ordStatus.status)

		case "6":
			c := make (chan OrderStatusReq)
			QueryWorkingOrder("venustech",c) // Get all working order
			wo := <- c

			c1 := make (chan OrderConfirmation)
			QueryOrderCancelReplace(wo.workingOrders[0].orderID,xid.New().String(),"venustech","1","2","250","4440","BZ","CME","201709","FUT", c1) // Replace the first working order
			ordStatus := <-c1
			fmt.Printf("Status: %s",ordStatus.status)

		case "9":
			break Loop

		default:
			err = fmt.Errorf("unknown action: '%v'", action)
		}

		if err != nil {
			fmt.Printf("%v\n", err)
		}
	}
}

func QueryAction() (string, error) {
	fmt.Println()
	fmt.Println("1) Get P&L")
	fmt.Println("2) Place new Order")
	fmt.Println("3) Get Working Order")
	fmt.Println("4) Market Data")
	fmt.Println("5) Cancel First working order")
	fmt.Println("6) Replace First working order")
	fmt.Println("9) Quit")
	fmt.Print("Action: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}
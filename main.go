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
			uan := TT_PAndLSOD(xid.New().String(), "TTORDFA222222")
			for _, uap := range uan.reports {
				fmt.Printf("%s \n", uap.product)
			}
		case "2":
			ordStatus := TT_NewOrderSingle(xid.New().String(), "venustech", "1", "2", "500", "4570", "BZ", "CME", "201709", "FUT")
			fmt.Printf("Status: %s \n", ordStatus.status)
		case "3":
			wo := TT_WorkingOrder("venustech")
			for _, order := range wo.workingOrders {
				fmt.Printf("%s \n", order.symbol)
			}
		case "4":
			mdr := TT_MarketDataRequest(xid.New().String(), "0", 0, "2", "BZ", "CME", "201709", "FUT")
			fmt.Printf("Price :%s \n", mdr.price)

		case "5":
			wo := TT_WorkingOrder("venustech") // get all working order

			ordStatus := TT_OrderCancel(xid.New().String(), wo.workingOrders[0].orderID) // Cancel the first working order
			fmt.Printf("Status: %s", ordStatus.status)
		case "6":
			wo := TT_WorkingOrder("venustech") // get all working order
			for i := range wo.workingOrders{
				ordStatus := TT_OrderCancel(xid.New().String(), wo.workingOrders[i].orderID) // Cancel the first working order
				fmt.Printf("Status: %s \n", ordStatus.status)
			}
		case "7":
			wo := TT_WorkingOrder("venustech") // get all working order
			// Replace/Edit the first working order
			ordStatus := TT_OrderCancelReplace(wo.workingOrders[0].orderID, xid.New().String(), "venustech", "1", "2", "250", "4440", "BZ", "CME", "201709", "FUT")
			fmt.Printf("Status: %s \n", ordStatus.status)

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
	fmt.Println("6) Cancel All working order")
	fmt.Println("7) Replace First working order")
	fmt.Println("9) Quit")
	fmt.Print("Action: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}

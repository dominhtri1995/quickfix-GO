package main

import (
	"fmt"
	"os"
	"bufio"
	"github.com/rs/xid"
	"os/exec"
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
			uan := TT_PAndLSOD(xid.New().String(), "venustech", "TTORDFA224222", "VENUSTECH3")
			if uan.Status == "rejected" {
				fmt.Printf("Error when getting UAN: %s", uan.Reason)
				continue
			}
			fmt.Printf("We have %d position in total for Account %s and accountgroup %s \n", len(uan.Reports), uan.Account, uan.AccountGroup)
			for _, uap := range uan.Reports {
				fmt.Printf("%s %s %s\n",uap.Side, uap.SecurityAltID,uap.Price)
			}
		case "2":
			ordStatus := TT_NewOrderSingle(xid.New().String(), "venustech", "1", "4", "50", "57.2","57", "IPE e-Brent", "ICE", "201709", "FUT", "1","0","0", "VENUSTECH3")
			if ordStatus.Status == "ok" {
				fmt.Println(ordStatus.Id)
				fmt.Printf("Order %s %s at %s Placed Successfully \n", ordStatus.Side, ordStatus.Symbol, ordStatus.Price)
			} else if ordStatus.Status == "rejected" {
				fmt.Printf("Order Rejected \n")
				if ordStatus.Reason != "" {
					fmt.Printf("Reason: %s", ordStatus.Reason)
				}
			}
		case "3":
			wo := TT_WorkingOrder("venustech", "VENUSTECH3")
			if wo.Status != "rejected" {
				for _, order := range wo.WorkingOrders {
					fmt.Printf("%s \n", order.Symbol)
				}
			} else {
				fmt.Printf("Error when getting Working Orders: %s", wo.Reason)
			}

		case "4":
			mdr := TT_MarketDataRequest(xid.New().String(), "0", 0, "2", "BZ", "CME", "201709", "FUT", "VENUSTECH3")
			fmt.Printf("Price :%s \n", mdr.Price)

		case "5":
			wo := TT_WorkingOrder("venustech", "VENUSTECH3") // get all working order

			ordStatus := TT_OrderCancel(xid.New().String(), wo.WorkingOrders[0].OrderID, "VENUSTECH3") // Cancel the first working order
			if ordStatus.Status == "ok" {
				fmt.Printf("Order  %s %s at %s Cancelled Successfully \n", ordStatus.Side, ordStatus.Symbol, ordStatus.Price)
			} else if ordStatus.Status == "rejected" {
				fmt.Printf("Order Cancelled Rejected \n")
				if ordStatus.Reason != "" {
					fmt.Printf("Reason: %s", ordStatus.Reason)
				}
			}
		case "6":
			wo := TT_WorkingOrder("venustech", "VENUSTECH3") // get all working order
			for i := range wo.WorkingOrders {
				ordStatus := TT_OrderCancel(xid.New().String(), wo.WorkingOrders[i].OrderID, "VENUSTECH3") // Cancel the first working order
				fmt.Printf("Status: %s \n", ordStatus.Status)
			}
		case "7":
			wo := TT_WorkingOrder("venustech", "VENUSTECH3") // get all working order
			// Replace/Edit the first working order
			order := wo.WorkingOrders[0]
			//Change Quantity of that working order to 962
			ordStatus := TT_OrderCancelReplace(order.OrderID, xid.New().String(), wo.Account, order.SideNum, order.OrdType, "962", order.Price,order.StopPrice, order.Symbol, order.Exchange, order.ProductMaturity, order.ProductType, order.TimeInForce,order.StrikePrice,order.PutOrCall, "VENUSTECH3")
			if ordStatus.Status == "ok" {
				fmt.Printf("Order Replaced Successfully: %s %s at %s \n", ordStatus.Side, ordStatus.Symbol, ordStatus.Price)
			} else if ordStatus.Status == "rejected" {
				fmt.Printf("Order Replaced Rejected \n")
				if ordStatus.Reason != "" {
					fmt.Printf("Reason: %s", ordStatus.Reason)
				}
			}
		case "8":
			//pass in parameters for filter, pass "" if do not want to use that criteria  //00A0IR00BZZ
			sdr := TT_QuerySecurityDefinitionRequest(xid.New().String(),"BZO","CME","","OPT","VENUSTECH3")
			if sdr.Status =="ok"{
				for _,security := range sdr.SecurityList {
					fmt.Printf("%s with TickValue: %f and TickSize %f %s %s %s\n",security.Symbol,security.TickValue,security.TickSize,security.SecurityAltID,security.ProductMaturity, security.Exchange)
				}
			}else{
				fmt.Println("error getting security definition")
				fmt.Printf("Reason: %s \n",sdr.Reason)
			}
		case "9":
			uan:= TT_Fills(xid.New().String(),"venustech3","VENUSTECH3")
			if uan.Status == "ok"{
				fmt.Println("Get Fill ok")
				for _, uap := range uan.Reports {
					fmt.Printf("Fill: %s %s at %s quant: %s\n",uap.Side, uap.SecurityAltID,uap.Price,uap.Quantity)
				}
			}else{
				fmt.Println("get fill failed")
			}

		case "99":
			break Loop
		case "19"://test ordertype
			numbers := []string{"5","B","J","O","Q","R"}
			for _,n := range numbers{
				ordStatus := TT_NewOrderSingle(xid.New().String(), "venustech", "1", "2", "50", "247825","2500222", "ES", "CME", "201709", "FUT", n,"0","0", "VENUSTECH3")
				if ordStatus.Status == "ok" {
					fmt.Println(ordStatus.Id)
					fmt.Printf("Order %s %s at %s Placed Successfully \n", ordStatus.Side, ordStatus.Symbol, ordStatus.Price)
				} else if ordStatus.Status == "rejected" {
					fmt.Printf("Order Rejected \n")
					if ordStatus.Reason != "" {
						fmt.Printf("Reason: %s\n", ordStatus.Reason)
					}
				}
			}


		default:
			cmd := exec.Command("clear") //linux only
			cmd.Stdout = os.Stdout
			cmd.Run()
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
	fmt.Println("8) Test Security Definition Request")
	fmt.Println("9) Get Fills")
	fmt.Println("99) Quit")
	fmt.Print("Action: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}

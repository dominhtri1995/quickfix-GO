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
			QueryPAndLSOD(xid.New().String(),"TTORDFA222222")
		case "2":
			QueryNewOrderSingle(xid.New().String(),"venustech","1","2","500","4570","BZ","CME","201709")
		case "3":
			QueryWorkingOrder("venustech")
		case "4":
			QueryMarketDataRequest(xid.New().String(),"0",0,"2","BZ","CME","201709")
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
	fmt.Println("4) Quit")
	fmt.Print("Action: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}
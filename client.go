package main

import (
	"fmt"
	"bufio"
	"os"
	"strings"
	"net/http"
	"io/ioutil"
	"time"
)

var apiUrl string

var userName string

func main() {
	apiUrl = "http://localhost:9999/"

	setUserName()

	go getMessagesC()

	showMenu()

	for true {
		InputHandler()
	}
}


func showMenu(){
	fmt.Println("")
	fmt.Println("--PLEASE SELECT THE DESIRED OPTION:\n")
	fmt.Println("  1. Create a chatroom.   Args: Name")
	fmt.Println("  2. List chatrooms.")
	fmt.Println("  3. Join existing chatroom.   Args: Name")
	fmt.Println("  4. Send Message to all joined chatrooms  Args: Message")
	fmt.Println("  5. Quit chatroom.    Args: Name")
	fmt.Println("  0. Show Menu")
	fmt.Println("")
	fmt.Println("  Example:  '3 chatroom2'")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("")

}

func InputHandler(){
	reader := bufio.NewReader(os.Stdin)

	for true {
		m, _ := reader.ReadString('\n')
		option, args := parseInput(m)
		if(option==""){
			fmt.Printf("Please select an option. Remember: 0 shows the menu:\n")
		}else{
			option = strings.Replace(option,"\n","",-1)
			switch option {

			//Show the menu
			case "0":
				showMenu()


			//Create chatroom
			case "1":
				if(args==""){
					fmt.Printf("Not Args found. Example: '1 NewChatRoom'")
				}else{
					fmt.Printf("%v\n", makeRequest("POST",apiUrl+"chatRooms/add/"+args))

					//message := Encode(args)
					//conn.Write([]byte("C/;"+args))

				}

			//List chatrooms
			case "2":

				fmt.Printf("%v\n", makeRequest("GET",apiUrl+"chatRooms/list"))



			//Join Existing chatroom
			case "3":
				if(args==""){
					fmt.Println("Not Args found. Example: '3 ExistingChatRoom'")
				}else{
					fmt.Printf("%v\n", makeRequest("POST",apiUrl+"chatRoom/"+args+"/join?user="+userName))

				}

			//Send message
			case "4":
				if(args==""){
					fmt.Println("Not Args found. Example: '4 hello everyone!'")
				}else{
					makeRequest("PUT",apiUrl+"message"+"?user="+userName+"&message="+args)
				}

			//Leave chatroom
			case "5":
				if(args==""){
					fmt.Println("Not Args found. Example: '5 ExistingChatRoom'")
				}else{
					fmt.Println(makeRequest("DELETE",apiUrl+"chatRoom/"+args+"/leave"+"?user="+userName))
				}

			default:

			}

		}
	}
}

func getMessagesC(){
	for{
		reply := makeRequest("GET",apiUrl+"userName/"+userName+"/messages")

		if reply != "" {
			fmt.Printf("%v\n", reply)
		}

		time.Sleep(300*time.Millisecond)

	}

}


func setUserName(){
	fmt.Println("Please set your username:")
	reader := bufio.NewReader(os.Stdin)
	userName, _ = reader.ReadString('\n')

	userName = strings.Replace(userName,"\n","",-1)

	fmt.Printf("%v\n", makeRequest("POST",apiUrl+"userName/"+userName))

}

func makeRequest(method, url string) string{
	client := &http.Client{}
	req,_:=http.NewRequest(method,url,nil)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

func parseInput(m string)(string, string){
	m = strings.Replace(m,"\n","",-1)
	splitted := strings.SplitN(m," ",2)
	if(len(splitted)>1){
		return Encode(splitted[0]),Encode(splitted[1])
	}
	if(len(splitted)==1){
		return Encode(splitted[0]),""
	}
	return "",""
}

func Decode(value string) (string) {
	var ENCODING_UNENCODED_TOKENS = []string{"%", ":", "[", "]", ",", "\"", " ","!","#","$","&","'","(",")","*","+","-",".","/"}
	var ENCODING_ENCODED_TOKENS = []string{"%25", "%3A", "%5B", "%5D", "%2C", "%22", "%20","%21","%23","%24","%26","%27","%28","%29","%2A","%2B","%2D","%2E","%2F"}
	return replace(ENCODING_ENCODED_TOKENS,ENCODING_UNENCODED_TOKENS, value)
}

func Encode(value string) (string) {
	var ENCODING_UNENCODED_TOKENS = []string{"%", ":", "[", "]", ",", "\"", " ","!","#","$","&","'","(",")","*","+","-",".","/"}
	var ENCODING_ENCODED_TOKENS = []string{"%25", "%3A", "%5B", "%5D", "%2C", "%22", "%20","%21","%23","%24","%26","%27","%28","%29","%2A","%2B","%2D","%2E","%2F"}
	return replace(ENCODING_UNENCODED_TOKENS, ENCODING_ENCODED_TOKENS, value)
}

func replace(fromTokens []string, toTokens []string, value string) (string) {
	for i:=0; i<len(fromTokens); i++ {
		value = strings.Replace(value, fromTokens[i], toTokens[i], -1)
	}
	return value;
}
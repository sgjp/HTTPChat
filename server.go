package main

import (
	"time"
	"github.com/gin-gonic/gin"
	//"html"
	"log"
	"fmt"
)


var cs *ChatServer

func main() {
	ListenAndServe("9999")

}

func ListenAndServe(port string){
	chatRooms := make(map[string]Chatroom)
	clients := make(map[string]Client)
	cs = new(ChatServer)
	//chatRooms["name"]=Chatroom{"name",nil,nil}

	cs.chatRooms = chatRooms
	cs.clients = clients



	go removeUnusedChatRooms(cs)

	router := gin.Default()

	router.POST("/chatRooms/add/:new", createChatRoom)

	router.GET("/chatRooms/list", listChatRooms)

	router.POST("/chatRoom/:cr/join", joinChatRoom)

	router.PUT("/message", messageChatRooms)

	router.DELETE("/chatRoom/:cr/leave", leaveChatRoom)

	router.POST("/userName/:user",setUsername)

	router.GET("/userName/:user/messages",getMessages)

	router.POST("/userName/:user/message",receiveMessage)

	router.Run(":"+port)

}


func createChatRoom(c *gin.Context) {
	msg := c.Param("new")


	if _, ok := cs.chatRooms[msg]; ok {
		c.String(200,"That chatRoom already exists")
		return
	}
	var clients []string
	var messages []Message

	chatRoom := Chatroom{msg,clients,messages}

	cs.chatRooms[msg] = chatRoom

	c.String(200,"That chatRoom was created !")

	return
}

func listChatRooms(c *gin.Context) {
	var content string
	for k:= range  cs.chatRooms{
		content += "*"+k+"\n"
	}
	c.String(200,content)
}

func joinChatRoom(c *gin.Context) {

	user := c.Query("user")
	chatroomName := c.Param("cr")
	var content string
	if _, ok := cs.clients[user]; ok {
		if _, ok2 := cs.chatRooms[chatroomName]; ok2 {
			chatRoom := cs.chatRooms[chatroomName]
			chatRoomClients := chatRoom.clients

			for k := range chatRoomClients{
				if chatRoomClients[k] == user{
					content = "You are already joined to this chatroom"
					c.String(200, content)
					return
				}
			}
			chatRoomClients = append(chatRoomClients,user)
			chatRoom.clients = chatRoomClients
			cs.chatRooms[chatroomName] = chatRoom

			content = "You joined the chatroom !\n"


			if len(cs.chatRooms[chatroomName].messages)==1{
				content += cs.chatRooms[chatroomName].messages[0].message

			}else if len(cs.chatRooms[chatroomName].messages)>1{
				var availableMessages string
				for _,msg:= range cs.chatRooms[chatroomName].messages{
					availableMessages = availableMessages +  msg.message +"\n"
				}

				content += availableMessages
			}
			c.String(200,content)
			return
		}else{
			c.String(200, "That chatRoom doesn't exist :(")
			return

		}
	}else{
		c.String(200, "That username doesn't exist!")
		return

	}

}

func messageChatRooms(c *gin.Context) {
	user := c.Query("user")
	msg := c.Query("message")
	go broadcastMessage(user,msg)
}

func leaveChatRoom(c *gin.Context) {
	user := c.Query("user")
	chatroomName := c.Param("cr")

	for k:= range  cs.chatRooms{
		if(k==chatroomName){
			//Go througn all the clients for the chatroom
			for i:= range cs.chatRooms[k].clients{
				//Is the user in this chatroom?
				if (cs.chatRooms[k].clients[i] == user) {
					chatRoom := cs.chatRooms[k]
					clients := chatRoom.clients
					clients = append(clients[:i],clients[i+1:]...)
					chatRoom.clients = clients
					cs.chatRooms[k] = chatRoom
					c.String(200 ,"You left the ChatRoom")
					return
				}
			}
		}
	}
	c.String(200, "You are not in the chatroom or it doesn't exist")
	return
}

func setUsername(c *gin.Context) {
	user := c.Param("user")


	if _, ok := cs.clients[user]; ok {
		c.String(200, "That username is already taken")
		return
	}

	client := new(Client)
	client.UserName = user
	cs.clients[client.UserName]=*client

	c.String(200,  "Welcome %v",client.UserName)
	return
}

func getMessages (c *gin.Context)   {
	user := c.Param("user")
	var content string

	if len(cs.clients[user].messagesToDeliver)==0{
		return
	}else if len(cs.clients[user].messagesToDeliver)==1{
		content = cs.clients[user].messagesToDeliver[0]
		client := cs.clients[user]

		client.messagesToDeliver = make([]string,0)

		cs.clients[user] = client
	}else{
		var availableMessages string
		for _,msg:= range cs.clients[user].messagesToDeliver{
			availableMessages = availableMessages +  msg
		}
		content = availableMessages

		client := cs.clients[user]

		client.messagesToDeliver = make([]string,0)

		cs.clients[user] = client
	}



	c.String(200, content)
	return

}

func receiveMessage(c *gin.Context) {
	msg := c.Query("messageContent")
	user := c.Param("user")
	log.Printf("Message Received: %v. CHATSERVER %v",msg,c)
	go broadcastMessage(user, msg)

	c.String(200,"")
	return
}

type ChatServer struct {
	chatRooms	map[string]Chatroom
	clients		map[string]Client
}

type Client struct {
	UserName string
	messagesToDeliver []string
}

type Chatroom struct{
	name string
	clients []string
	messages []Message
}

type Message struct{
	message string
	date time.Time
}


func removeUnusedChatRooms(cs *ChatServer){
	for{
		time.Sleep(1000* time.Millisecond)
		currentDate := time.Now
		var newChatRooms = make(map[string]Chatroom)
		var deletedFlag bool
		for k := range cs.chatRooms {
			var lastMessage Message
			//Go through all the messages for the chatroom
			for i:= range cs.chatRooms[k].messages{
				if lastMessage.message == ""{
					lastMessage=cs.chatRooms[k].messages[i]
				}else if lastMessage.date.Before(cs.chatRooms[k].messages[i].date){
					lastMessage=cs.chatRooms[k].messages[i]
				}else{
					lastMessage = lastMessage
				}


			}
			//if the chatroom has been used in the last 7 days or never been used, keep it
			if currentDate().AddDate(0,0,-7).Before(lastMessage.date) || lastMessage.message==""{
				//log.Printf("Keeping ChatRoom: %v",cs.chatRooms[k].name)
				newChatRooms[cs.chatRooms[k].name]=cs.chatRooms[k]
				deletedFlag = true
			}

		}
		if deletedFlag{
			cs.chatRooms = newChatRooms
		}


	}


}

func broadcastMessage(user, message string){

	currentTime := time.Now()
	message = currentTime.Format("[Jan _2 15:04:05]")+" " + user + ": " + message

	chatRoomsToBroadcast := make([]string,0)

	for k, _ := range cs.chatRooms {
		flagContinue := true
		fmt.Printf("CRs: %v",cs.chatRooms)
		for i:= range cs.chatRooms[k].clients{
			if cs.chatRooms[k].clients[i]==user{
				chatRoomsToBroadcast = append(chatRoomsToBroadcast,k)
				flagContinue = false
				addMessageToChatroom(k,message)
				continue
			}
		}
		if !flagContinue{
			continue
		}
	}

	for k := range chatRoomsToBroadcast {
		//log.Printf("CHATROOM joined %v",c.chatRooms[k])
		chatRoomName := chatRoomsToBroadcast[k]
		for i:= range cs.chatRooms[chatRoomName].clients{
			if cs.chatRooms[chatRoomName].clients[i]!=user{
				messagesToDeliver := cs.clients[cs.chatRooms[chatRoomName].clients[i]].messagesToDeliver
				messagesToDeliver = append(messagesToDeliver,message)

				client := cs.clients[cs.chatRooms[chatRoomName].clients[i]]
				client.messagesToDeliver = messagesToDeliver

				cs.clients[cs.chatRooms[chatRoomName].clients[i]]= client
				//append(c.chatRooms[k].clients[i].messagesToDeliver,message)
			}
		}

	}
}

func addMessageToChatroom(name, message string){
	chatRoom := cs.chatRooms[name]
	newMessage := Message{message:message,date:time.Now()}
	chatRoom.messages = append(chatRoom.messages,newMessage)
	cs.chatRooms[name] = chatRoom
}
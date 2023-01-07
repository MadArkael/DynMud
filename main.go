package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

var clientIn chan ClientInput
var clientOut chan ClientOutput
var consoleIn chan string
var consoleOut chan string
var W *World

const (
	port     int    = 8080
	textWrap int    = 70
	BLACK    string = "\u001b[30m"
	RED      string = "\u001b[31m"
	GREEN    string = "\u001b[32m"
	YELLOW   string = "\u001b[33m"
	BLUE     string = "\u001b[34m"
	MAGENTA  string = "\u001b[35m"
	CYAN     string = "\u001b[36m"
	WHITE    string = "\u001b[37m"
	RESET    string = "\u001b[0m"
	NWLN     string = "\r\n"
)

type Room struct {
	name    string
	nColor  string
	desc    string
	dColor  string
	char    string
	clients []*Client
	z       int
	x       int
	y       int
}

type World struct {
	rooms   []*Room
	roomMap map[int]map[int]map[int]*Room //z,x,y,Room
}

type ClientInput struct {
	cl  *Client
	evt interface{}
}

type ClientOutput struct {
	cl  *Client
	evt interface{}
}

type InputEvent struct {
	input string
}

type OutputEvent struct {
	output string
}

type DisconnectEvent struct {
}

type ConnectEvent struct {
}

type Client struct {
	conn net.Conn
	name string
	z    int
	x    int
	y    int
	room *Room
}

func initRooms() []*Room {
	rooms := []*Room{
		{name: "The Plaza", nColor: BLUE, desc: "    The plaza is a hub of activity, with people bustling about and vendors shouting out their wares. The fountain at the center of the plaza gurgles and sprays water into the air, providing a refreshing respite from the hot sun. The energy of the city seems to radiate from this one central location, making it a true heart of the community.", dColor: RED, z: 0, x: 0, y: 0}, //central
		{name: "North of the Plaza", nColor: BLUE, desc: "    The street to the north is bustling with activity, as merchants hawk their wares and peasants haggle for goods. The smell of roasting meat and freshly baked bread fills the air.", dColor: RED, z: 0, x: 0, y: 1},
		{z: 0, x: 0, y: 2}, //naryas
		{z: 0, x: 0, y: 3},
		{z: 0, x: 0, y: 4},
		{z: 0, x: 0, y: 5}, //gate fire
		{name: "South of the Plaza", nColor: BLUE, desc: "    The street to the south is a busy thoroughfare, with knights on horseback and peasants on foot traveling to and fro. The sound of clanging swords and armor can be heard from the nearby training grounds.", dColor: RED, z: 0, x: 0, y: -1},
		{z: 0, x: 0, y: -2},
		{z: 0, x: 0, y: -3},
		{z: 0, x: 0, y: -4},
		{z: 0, x: 0, y: -5}, //gate air
		{z: 0, x: -5, y: 0}, //gate water
		{z: 0, x: -4, y: 0},
		{z: 0, x: -3, y: 0},
		{z: 0, x: -2, y: 0},
		{name: "West of the Plaza", nColor: BLUE, desc: "    The street to the west is a mix of noble manors and humble homes. Children play in the dirt streets, and the occasional chicken or goat can be seen wandering about.", dColor: RED, z: 0, x: -1, y: 0},
		{name: "East of the Plaza", nColor: BLUE, desc: "    The street to the east is more peaceful, with quaint cottages and gardens dotted along the way. The occasional horse-drawn carriage clatters by, kicking up a trail of dust.", dColor: RED, z: 0, x: 1, y: 0},
		{z: 0, x: 2, y: 0},
		{z: 0, x: 3, y: 0},
		{z: 0, x: 4, y: 0},
		{z: 0, x: 5, y: 0}, //gate earth
		{z: 0, x: -2, y: 1},
		{z: 0, x: -2, y: 2},
		{z: 0, x: -1, y: 2},
		{z: 0, x: 2, y: 1},
		{z: 0, x: 2, y: 2},
		{z: 0, x: 1, y: 2},
		{z: 0, x: 2, y: -1},
		{z: 0, x: 2, y: -2},
		{z: 0, x: 1, y: -2},
		{z: 0, x: -2, y: -1},
		{z: 0, x: -2, y: -2}, //arn
		{z: 0, x: -1, y: -2},
		{z: 0, x: -2, y: 3},  //Sarse
		{z: 0, x: -3, y: -1}, //waysl
		{z: 0, x: 3, y: -2},  //jansen
		{z: 0, x: 1, y: -3},  //teller
		{z: 0, x: -3, y: -2}, //Farath
		{z: 0, x: 1, y: 3},   //Murlain
		{z: 0, x: 3, y: 2},   //Alentys
		{z: 0, x: -1, y: 3},  //farvik
		{z: 0, x: -3, y: 1},  //borash
	}
	return rooms
}

func addRooms(rooms []*Room, roomsMap map[int]map[int]map[int]*Room) {
	for _, room := range rooms {
		if roomsMap[room.z] == nil {
			roomsMap[room.z] = make(map[int]map[int]*Room)
		}
		if roomsMap[room.z][room.y] == nil {
			roomsMap[room.z][room.y] = make(map[int]*Room)
		}
		roomsMap[room.z][room.y][room.x] = room
	}
}

// returns z, x, y
func (c *Client) getLocation() (int, int, int) {
	return c.z, c.y, c.x
}

func (r *Room) getLocation() (int, int, int) {
	return r.z, r.y, r.x
}

func (c *Client) getPrompt() string {
	z, y, x := c.getLocation()
	n := []int{z, y + 1, x}
	s := []int{z, y - 1, x}
	e := []int{z, y, x + 1}
	w := []int{z, y, x - 1}
	u := []int{z + 1, y, x}
	d := []int{z - 1, y, x}
	ne := []int{z, y + 1, x + 1}
	nw := []int{z, y + 1, x - 1}
	se := []int{z, y - 1, x + 1}
	sw := []int{z, y - 1, x - 1}
	prompt := "E: "
	if _, ok := W.roomMap[ne[0]][ne[1]][ne[2]]; ok {
		prompt += "Ne"
	}
	if _, ok := W.roomMap[nw[0]][nw[1]][nw[2]]; ok {
		prompt += "Nw"
	}
	if _, ok := W.roomMap[se[0]][se[1]][se[2]]; ok {
		prompt += "Se"
	}
	if _, ok := W.roomMap[sw[0]][sw[1]][sw[2]]; ok {
		prompt += "Sw"
	}
	if _, ok := W.roomMap[n[0]][n[1]][n[2]]; ok {
		prompt += "N"
	}
	if _, ok := W.roomMap[s[0]][s[1]][s[2]]; ok {
		prompt += "S"
	}
	if _, ok := W.roomMap[e[0]][e[1]][e[2]]; ok {
		prompt += "E"
	}
	if _, ok := W.roomMap[w[0]][w[1]][w[2]]; ok {
		prompt += "W"
	}
	if _, ok := W.roomMap[u[0]][u[1]][u[2]]; ok {
		prompt += "U"
	}
	if _, ok := W.roomMap[d[0]][d[1]][d[2]]; ok {
		prompt += "D"
	}
	return prompt
}

func removeClientFromSlice(c *Client, cs []*Client) []*Client {
	var temp []*Client
	for n, cl := range cs {
		if cl == c {
			temp = append(cs[:n], cs[n+1:]...)
		}
	}
	return temp
}

func (c *Client) broadcastLeaving(dir string) {
	for _, cl := range c.room.clients {
		if cl != c {
			clientOut <- ClientOutput{cl, &OutputEvent{fmt.Sprintf(CYAN+"%s"+RESET+" heads %s.", c.name, dir)}}
		}
	}
}

func opposite(dir string) string {
	s := ""
	switch dir {
	case "n":
		s = "south"
	case "s":
		s = "north"
	case "e":
		s = "west"
	case "w":
		s = "east"
	case "nw":
		s = "south east"
	case "ne":
		s = "south west"
	case "se":
		s = "north west"
	case "sw":
		s = "north east"
	case "u":
		s = "below"
	case "d":
		s = "above"
	}
	return s
}

func (c *Client) broadcastArriving(from string) {
	clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You head %s.\r\n%s", from, c.room.getDisplay(c))}}
	for _, cl := range c.room.clients {
		if cl != c {
			clientOut <- ClientOutput{cl, &OutputEvent{fmt.Sprintf(CYAN+"%s"+RESET+" arrives from the %s.", c.name, opposite(from))}}
		}
	}
}

func (c *Client) updateClientLocation(z int, y int, x int, dir string) {
	c.broadcastLeaving(dir)
	c.room.clients = removeClientFromSlice(c, c.room.clients)
	c.room = getRoomByCoords(z, y, x)
	c.room.clients = append(c.room.clients, c)
	c.z = z
	c.y = y
	c.x = x
	c.broadcastArriving(dir)
}

func (c *Client) move(dir string) {
	z, y, x := c.getLocation()
	n := []int{z, y + 1, x}
	s := []int{z, y - 1, x}
	e := []int{z, y, x + 1}
	w := []int{z, y, x - 1}
	u := []int{z + 1, y, x}
	d := []int{z - 1, y, x}
	ne := []int{z, y + 1, x + 1}
	nw := []int{z, y + 1, x - 1}
	se := []int{z, y - 1, x + 1}
	sw := []int{z, y - 1, x - 1}
	switch dir {
	case "ne":
		if r, ok := W.roomMap[ne[0]][ne[1]][ne[2]]; ok {
			c.updateClientLocation(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "nw":
		if r, ok := W.roomMap[nw[0]][nw[1]][nw[2]]; ok {
			c.updateClientLocation(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "se":
		if r, ok := W.roomMap[se[0]][se[1]][se[2]]; ok {
			c.updateClientLocation(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "sw":
		if r, ok := W.roomMap[sw[0]][sw[1]][sw[2]]; ok {
			c.updateClientLocation(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "n":
		if r, ok := W.roomMap[n[0]][n[1]][n[2]]; ok {
			c.updateClientLocation(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "s":
		if r, ok := W.roomMap[s[0]][s[1]][s[2]]; ok {
			c.updateClientLocation(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "e":
		if r, ok := W.roomMap[e[0]][e[1]][e[2]]; ok {
			c.updateClientLocation(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "w":
		if r, ok := W.roomMap[w[0]][w[1]][w[2]]; ok {
			c.updateClientLocation(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "u":
		if r, ok := W.roomMap[u[0]][u[1]][u[2]]; ok {
			c.updateClientLocation(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "d":
		if r, ok := W.roomMap[d[0]][d[1]][d[2]]; ok {
			c.updateClientLocation(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	}
}

// take strng, color it with color, and wrap it neatly at lng length max.
func textWrapString(color string, strng string, lng int) string {
	s := ""
	x := lng
	oldx := 0
	lenDesc := len(strng)
	for x < lenDesc {
		for strng[x-1:x] != " " {
			x--
		}
		s += color + strng[oldx:x] + RESET + NWLN
		oldx = x
		if x+lng >= lenDesc {
			s += color + strng[oldx:] + RESET
			break
		}
		x += lng
	}
	return s
}

func (r *Room) dispOtherClients(cl *Client) string {
	str := ""
	for _, c := range r.clients {
		if c != cl {
			str += NWLN + "    " + CYAN + c.name + RESET + " is standing here."
		}
	}
	return str
}

func buildMap(cl *Client) string {
	zVar, yVar, xVar := cl.getLocation()
	textMap := ""
	for y := yVar + 5; y > yVar-6; y-- {
		for x := xVar - 5; x < xVar+6; x++ {
			if r, ok := W.roomMap[zVar][y][x]; ok {
				if y == yVar && x == xVar {
					textMap = textMap + CYAN + " R" + RESET
				} else {
					if len(r.clients) != 0 {
						textMap = textMap + RED + " R" + RESET
					} else {
						textMap = textMap + " R"
					}
				}
			} else {
				textMap = textMap + " -"
			}
		}
		if y == yVar-5 {
			return textMap
		}
		textMap = textMap + "\r\n"
	}
	return textMap
}

func (r *Room) getDisplay(cl *Client) string {
	str := ""
	lines := strings.Split(buildMap(cl), "\r\n")
	lineD := strings.Split(r.nColor+r.name+RESET+NWLN+textWrapString(r.dColor, r.desc, textWrap-25)+r.dispOtherClients(cl), "\r\n")

	lLines := len(lines)
	lDesc := len(lineD)
	if lLines > lDesc {
		dif := lLines - lDesc
		for i := 0; i < lLines; i++ {
			if i < dif {
				str += lines[i] + " | " + NWLN
			} else {
				str += lines[i] + " | " + lineD[i-dif] + NWLN
			}
		}
	}
	if lLines == lDesc {
		for i := 0; i < lLines; i++ {
			str += lines[i] + " | " + lineD[i] + NWLN
		}
	}
	if lLines < lDesc {
		dif := lDesc - lLines
		for i := 0; i < lDesc; i++ {
			if i < dif {
				str += "                      " + " | " + lineD[i] + NWLN
			} else {
				str += lines[i-dif] + " | " + lineD[i] + NWLN
			}
		}
	}
	/*for i := 0; i < maxLen; i++ {
		if i < lLines && i < lDesc {
			str += lines[i] + " | " + lineD[i] + NWLN
		} else {
			if i < lLines {
				str += lines[i] + " | " + NWLN
			}
			if i < lDesc {
				str += "                      " + " | " + lineD[i] + NWLN
			}
		}
	}*/
	//str += r.nColor + r.name + RESET + NWLN
	//str += textWrapString(r.dColor, r.desc)
	//str += r.dispOtherClients(cl)
	return str[:len(str)-2]
}

func (c *Client) writeLnF(msg string, args ...interface{}) {
	fmt.Fprintf(c.conn, msg+"\r\n", args...)
}

func getRoomByCoords(z int, y int, x int) *Room {
	if r, ok := W.roomMap[z][y][x]; ok {
		return r
	}
	consoleOut <- fmt.Sprintf("No such room at coord: Z:%d, Y:%d, X:%d", z, y, x)
	return nil
}

func clientHandler(chI chan ClientInput, chO chan ClientOutput, conn net.Conn) {
	reader := bufio.NewReader(conn)
	fmt.Fprintf(conn, "What's your name?")
	name, er := reader.ReadString('\n')
	if er != nil {
		consoleOut <- "Error receiving name line from client."
		conn.Close()
		return
	}

	cl := &Client{conn: conn, name: name[:len(name)-2], x: 0, y: 0, z: 0, room: getRoomByCoords(0, 0, 0)}
	chO <- ClientOutput{cl, &OutputEvent{fmt.Sprintf("Thanks %s!\r\n"+cl.room.getDisplay(cl), cl.name)}}
	cl.room.clients = append(cl.room.clients, cl)
	chO <- ClientOutput{cl, &ConnectEvent{}}
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			consoleOut <- "Error receiving line from client."
			clientOut <- ClientOutput{cl, &DisconnectEvent{}}
			conn.Close()
			return
		}

		consoleOut <- fmt.Sprintf("%s: %s", cl.name, input[:len(input)-2])
		chI <- ClientInput{cl, &InputEvent{input[:len(input)-2]}}
	}
}

func startListener(chI chan ClientInput, chO chan ClientOutput) error {

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		consoleOut <- "Error accepting connection at listener."
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			consoleOut <- "Error handing off the connection to the client."
			conn.Close()
			return err
		}
		go clientHandler(chI, chO, conn)

	}
}

func startCIn(chI <-chan string) {

}

func startCOut(chO <-chan string) {
	for s := range chO {
		fmt.Println(s)
	}
}

func startIn(chI <-chan ClientInput, chO chan ClientOutput) {
	for in := range chI {
		switch e := in.evt.(type) {
		case *InputEvent:
			switch e.input {
			case "l":
				chO <- ClientOutput{in.cl, &OutputEvent{in.cl.room.getDisplay(in.cl)}}
			case "coords":
				chO <- ClientOutput{in.cl, &OutputEvent{fmt.Sprintf("Char coords: Z:%d Y:%d X%d\r\nRoom coords: Z:%d Y:%d X%d", in.cl.z, in.cl.y, in.cl.x, in.cl.room.z, in.cl.room.y, in.cl.room.x)}}
				fmt.Println(in.cl.room.clients)
			case "map":
				chO <- ClientOutput{in.cl, &OutputEvent{buildMap(in.cl)}}
			case "ne", "nw", "se", "sw", "n", "e", "s", "w", "u", "d":
				in.cl.move(e.input)
			default:
				chO <- ClientOutput{in.cl, &OutputEvent{fmt.Sprintf("You: %s", e.input)}}
				for _, c := range in.cl.room.clients {
					if c != in.cl {
						chO <- ClientOutput{c, &OutputEvent{fmt.Sprintf("%s: %s", in.cl.name, e.input)}}
					}
				}
			}
		}
	}
}

func startOut(chO <-chan ClientOutput) {
	for out := range chO {
		switch e := out.evt.(type) {
		case *OutputEvent:
			out.cl.writeLnF(e.output)
		case *DisconnectEvent:
			for _, c := range out.cl.room.clients {
				if c != out.cl {
					c.writeLnF(CYAN+"%s"+RESET+" has disconnected.", out.cl.name)
				}
			}
			out.cl.room.clients = removeClientFromSlice(out.cl, out.cl.room.clients)
		case *ConnectEvent:
			for _, c := range out.cl.room.clients {
				if c != out.cl {
					c.writeLnF(CYAN+"%s"+RESET+" has connected.", out.cl.name)
				}
			}
		}
		out.cl.writeLnF(out.cl.getPrompt())
	}
}

func main() {
	clientIn = make(chan ClientInput)
	clientOut = make(chan ClientOutput)
	consoleIn = make(chan string)
	consoleOut = make(chan string)
	W = &World{rooms: initRooms(), roomMap: make(map[int]map[int]map[int]*Room)}
	addRooms(W.rooms, W.roomMap)
	go startIn(clientIn, clientOut)
	go startOut(clientOut)
	go startCIn(consoleIn)
	go startCOut(consoleOut)

	if err := startListener(clientIn, clientOut); err != nil {

		consoleOut <- err.Error()
	}
}

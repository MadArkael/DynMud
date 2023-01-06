package main

import (
	"bufio"
	"fmt"
	"math"
	"net"
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
	desc    string
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
		{name: "The Plaza", desc: "    The plaza is a hub of activity, with people bustling about and vendors shouting out their wares. The smells of delicious street food fill the air, and the sound of music and laughter can be heard everywhere. The fountain at the center of the plaza gurgles and sprays water into the air, providing a refreshing respite from the hot sun. Street performers and artists line the edges of the plaza, offering a constant stream of entertainment. The energy of the city seems to radiate from this one central location, making it a true heart of the community.", z: 0, x: 0, y: 0}, //central
		{z: 0, x: 0, y: 1},
		{z: 0, x: 0, y: 2}, //naryas
		{z: 0, x: 0, y: 3},
		{z: 0, x: 0, y: 4},
		{z: 0, x: 0, y: 5}, //gate fire
		{z: 0, x: 0, y: -1},
		{z: 0, x: 0, y: -2},
		{z: 0, x: 0, y: -3},
		{z: 0, x: 0, y: -4},
		{z: 0, x: 0, y: -5}, //gate air
		{z: 0, x: -5, y: 0}, //gate water
		{z: 0, x: -4, y: 0},
		{z: 0, x: -3, y: 0},
		{z: 0, x: -2, y: 0},
		{z: 0, x: -1, y: 0},
		{z: 0, x: 1, y: 0},
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
		textMap = textMap + "\r\n"
	}
	return textMap
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

func (c *Client) update(z int, y int, x int, dir string) {
	c.room.clients = removeClientFromSlice(c, c.room.clients)
	c.room = getRoomByCoords(z, y, x)
	c.room.clients = append(c.room.clients, c)
	c.z = z
	c.y = y
	c.x = x
	clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You head %s.\r\n%s", dir, c.room.getDisplay(c))}}
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
			c.update(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "nw":
		if r, ok := W.roomMap[nw[0]][nw[1]][nw[2]]; ok {
			c.update(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "se":
		if r, ok := W.roomMap[se[0]][se[1]][se[2]]; ok {
			c.update(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "sw":
		if r, ok := W.roomMap[sw[0]][sw[1]][sw[2]]; ok {
			c.update(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "n":
		if r, ok := W.roomMap[n[0]][n[1]][n[2]]; ok {
			c.update(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "s":
		if r, ok := W.roomMap[s[0]][s[1]][s[2]]; ok {
			c.update(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "e":
		if r, ok := W.roomMap[e[0]][e[1]][e[2]]; ok {
			c.update(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "w":
		if r, ok := W.roomMap[w[0]][w[1]][w[2]]; ok {
			c.update(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "u":
		if r, ok := W.roomMap[u[0]][u[1]][u[2]]; ok {
			c.update(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	case "d":
		if r, ok := W.roomMap[d[0]][d[1]][d[2]]; ok {
			c.update(r.z, r.y, r.x, dir)
		} else {
			clientOut <- ClientOutput{c, &OutputEvent{fmt.Sprintf("You cant go %s.", dir)}}
		}
	}
}

func (r *Room) getDisplay(cl *Client) string {
	lines := math.Ceil(float64(len(r.desc)) / float64(textWrap))
	str := BLUE + r.name + RESET + NWLN
	pos1 := 0
	pos2 := textWrap
	breakFor := false
	for x := 0; x < int(lines); x++ {
		s := r.desc[pos1:pos2]
		pos1 = pos2
		pos2 = pos2 + textWrap

		if pos2 > len(r.desc) {
			pos2 = len(r.desc)
			breakFor = true
		}
		if !breakFor {
			for r.desc[pos2-1:pos2] != " " {
				pos2--
			}
		}
		if pos1 < len(r.desc)-70 {
			x--
		}
		str += RED + s + NWLN + RESET
		if pos1 == len(r.desc) {
			str = str[:len(str)-6] + RESET
			break
		}
	}
	for _, c := range r.clients {
		if c != cl {
			str += NWLN + "    " + c.name + " is standing here."
		}
	}

	return str
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
		return
	}

	cl := &Client{conn: conn, name: name[:len(name)-2], x: 0, y: 0, z: 0, room: getRoomByCoords(0, 0, 0)}
	chO <- ClientOutput{cl, &OutputEvent{fmt.Sprintf("Thanks %s!\r\n"+cl.room.getDisplay(cl), cl.name)}}
	cl.room.clients = append(cl.room.clients, cl)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			consoleOut <- "Error receiving line from client."
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

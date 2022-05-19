package main

import (
	"unicode/utf8"
	"flag"
	"fmt"
	"strings"
	"net"
	"github.com/bwmarrin/discordgo"
)

var (
	Token string
	Port int
	UDPSocket *net.UDPConn
	ChannelID string
	ServerList []string
	ServerListArg string
	ServerAddressList []net.UDPAddr
	QuakeSymbolMap = []rune{
		0, '#',  '#',  '#',  '#',  '.',  '#',  '#',
		'#',  9,    10,   '#',  ' ',  13,   '.',  '.',
		'[',  ']',  '0',  '1',  '2',  '3',  '4',  '5',
		'6',  '7',  '8',  '9',  '.',  '<',  '=',  '>',
		' ',  '!',  '"',  '#',  '$',  '%',  '&',  '\'',
		'(',  ')',  '*',  '+',  ',',  '-',  '.',  '/',
		'0',  '1',  '2',  '3',  '4',  '5',  '6',  '7',
		'8',  '9',  ':',  ';',  '<',  '=',  '>',  '?',
		'@',  'A',  'B',  'C',  'D',  'E',  'F',  'G',
		'H',  'I',  'J',  'K',  'L',  'M',  'N',  'O',
		'P',  'Q',  'R',  'S',  'T',  'U',  'V',  'W',
		'X',  'Y',  'Z',  '[',  '\\', ']',  '^',  '_',
		'`',  'a',  'b',  'c',  'd',  'e',  'f',  'g',
		'h',  'i',  'j',  'k',  'l',  'm',  'n',  'o',
		'p',  'q',  'r',  's',  't',  'u',  'v',  'w',
		'x',  'y',  'z',  '{',  '|',  '}',  '~',  '<',
		'<',  '=',  '>',  '#',  '#',  '.',  '#',  '#',
		'#',  '#',  ' ',  '#',  ' ',  '>',  '.',  '.',
		'[',  ']',  '0',  '1',  '2',  '3',  '4',  '5',
		'6',  '7',  '8',  '9',  '.',  '<',  '=',  '>',
		' ',  '!',  '"',  '#',  '$',  '%',  '&',  '\'',
		'(',  ')',  '*',  '+',  ',',  '-',  '.',  '/',
		'0',  '1',  '2',  '3',  '4',  '5',  '6',  '7',
		'8',  '9',  ':',  ';',  '<',  '=',  '>',  '?',
		'@',  'A',  'B',  'C',  'D',  'E',  'F',  'G',
		'H',  'I',  'J',  'K',  'L',  'M',  'N',  'O',
		'P',  'Q',  'R',  'S',  'T',  'U',  'V',  'W',
		'X',  'Y',  'Z',  '[',  '\\', ']',  '^',  '_',
		'`',  'a',  'b',  'c',  'd',  'e',  'f',  'g',
		'h',  'i',  'j',  'k',  'l',  'm',  'n',  'o',
		'p',  'q',  'r',  's',  't',  'u',  'v',  'w',
		'x',  'y',  'z',  '{',  '|',  '}',  '~',  '<',
	}
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&ChannelID, "c", "", "Channel id")
	flag.StringVar(&ServerListArg, "s", "", "Server list")
	flag.IntVar(&Port, "p", 10000, "UDP chat port")
	flag.Parse()
}

func main() {
	// Create a new Discord session using the provided bot token.
	ServerList = strings.Split(ServerListArg, ",")
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("discordgo New:", err)
		return
	}
	listen_addr := net.UDPAddr{
		Port: Port,
		IP:   net.ParseIP("0.0.0.0"),
	}
	UDPSocket_temp, udp_err := net.ListenUDP("udp", &listen_addr)
	UDPSocket = UDPSocket_temp
	if udp_err != nil {
		fmt.Println("net ListenPacket:", err)
		return
	}
	defer UDPSocket.Close()
	dg.AddHandler(messageHandle)
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)
	err = dg.Open()
	if err != nil {
		fmt.Println("discordgo Open:", err)
		return
	}
	buf := make([]byte, 2048)
	u8buffer := make([]byte, 7)
	for {
		n, addr, err := UDPSocket.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("net ReadFrom:", err)
			break
		}
		skip := true
		for _, server_addr := range ServerAddressList {
			if server_addr.String() == addr.String() {
				skip = false
				break
			}
		}
		if skip {
			addr_string := addr.String()
			for _, server := range ServerList {
				if (addr_string == server) {
					skip = false
					fmt.Println("Append address " + server)
					ServerAddressList = append(ServerAddressList, *addr)
					break
				}
			}
		}
		if skip {
			fmt.Println("Address '" + addr.String() + "' not found")
			continue
		}
		if (n > 4 && buf[0] == '\377' && buf[1] == '\377' && buf[2] == '\377' && buf[3] == '\377') {
			for _, server_addr := range ServerAddressList {
				if addr.String() != server_addr.String() {
					UDPSocket.WriteTo(buf, addr)
				}
			}
			s := string(buf[4:n])
			fmt.Println(s)
			if (s[0:19] == "extResponse udpchat") {
				s1 := []byte(s[19:])
				outstr := ""
				for len(s1) > 0 {
					r, r_size := utf8.DecodeRune(s1);
					s1 = s1[r_size:]
					if r >= 0xE000 && r < 0xE100 {
						r -= 0xE000;
					}
					if (r > 0) {
						if r < 256 {
							r = QuakeSymbolMap[r]
						}
						r_size = utf8.EncodeRune(u8buffer, r)
						outstr += string(u8buffer[:r_size])
					}
				}
				dg.ChannelMessageSend(ChannelID, outstr);
			} else {
				fmt.Println("Incorrect packet")
			}
		} else {
			fmt.Println("Incorrect packet")
		}
	}
	dg.Close()
}

func messageHandle(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if (m.ChannelID != ChannelID) {
		fmt.Println("Wrong channel id:", m.ChannelID)
		return
	}
	b := []byte("\377\377\377\377extResponse udpchat " + m.Author.Username + "@discord: " + m.Content)
	for _, addr := range ServerAddressList {
		_, err := UDPSocket.WriteToUDP(b, &addr)
		if err != nil {
			fmt.Println("net UDPConn WriteTo:", err)
		}
	}
}

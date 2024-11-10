package smtpx

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

// SMTPServer represents a basic SMTP server
type SMTPServer struct {
	Addr string
}

// Handle incoming connections
func (s *SMTPServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	writer.WriteString("220 SMTPX Server\r\n")
	writer.Flush()

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from connection: %v\n", err)
			return
		}

		log.Printf("Received: %s", line)
		cmd := strings.TrimSpace(line)

		if strings.HasPrefix(strings.ToUpper(cmd), "HELO") {
			writer.WriteString("250 Hello\r\n")
		} else if strings.HasPrefix(strings.ToUpper(cmd), "MAIL FROM:") {
			writer.WriteString("250 OK\r\n")
		} else if strings.HasPrefix(strings.ToUpper(cmd), "RCPT TO:") {
			writer.WriteString("250 OK\r\n")
		} else if strings.HasPrefix(strings.ToUpper(cmd), "DATA") {
			writer.WriteString("354 End data with <CR><LF>.<CR><LF>\r\n")
			writer.Flush()

			data := ""
			for {
				line, err = reader.ReadString('\n')
				if err != nil {
					log.Printf("Error reading from connection: %v\n", err)
					return
				}

				if line == ".\r\n" {
					break
				}
				data += line
			}

			log.Printf("Received mail data: %s", data)
			writer.WriteString("250 OK\r\n")
		} else if strings.HasPrefix(strings.ToUpper(cmd), "QUIT") {
			writer.WriteString("221 Bye\r\n")
			writer.Flush()
			return
		} else {
			writer.WriteString("500 Unrecognized command\r\n")
		}

		writer.Flush()
	}
}

// ListenAndServe starts the SMTP server
func (s *SMTPServer) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", s.Addr, err)
	}
	defer listener.Close()

	log.Printf("SMTP Server started on %s", s.Addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func main() {
	server := &SMTPServer{
		Addr: ":2525", // Use port 2525 for non-privileged access
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error starting SMTP server: %v", err)
	}
}

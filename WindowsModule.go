package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"time"
)

var socketC net.Conn
var fsalir bool

func loginUser() bool {
	fmt.Println("Digite el nombre de usuario:")
	var user string
	fmt.Scanf("%s", &user)
	tcpAddress, err := net.ResolveTCPAddr("tcp4", "10.0.2.15:1306")
	fmt.Println("Client# conectando con  ...", tcpAddress.IP)
	socketC, err = net.DialTCP("tcp", nil, tcpAddress)
	fmt.Println("Client# conectado con ..", socketC.RemoteAddr())

	err = gob.NewEncoder(socketC).Encode(&user)
	if err != nil {
		fmt.Println("Error al enviar el nombre de usuario:", err)
		return false
	}

	var usuariosPermitidos []string
	err = gob.NewDecoder(socketC).Decode(&usuariosPermitidos)
	if err != nil {
		fmt.Println("Error al recibir la lista de usuarios permitidos:", err)
		return false
	}

	permitido := false
	for _, u := range usuariosPermitidos {
		if u == user {
			permitido = true
			break
		}
	}
	return permitido
}

func authenticatePass() {
	for i := 1; i < 4; i++ {

		fmt.Println("\n Digite la contraseña:")
		var password string
		fmt.Scanf("%s", &password)

		err := gob.NewEncoder(socketC).Encode(&password)

		if err != nil {
			break
		}

		var autenticado bool
		err = gob.NewDecoder(socketC).Decode(&autenticado)
		if err != nil {
			fmt.Println("Error al recibir la autenticación:", err)
			break
		}

		if autenticado {
			fmt.Println("Login Correcto")
			break
		} else {
			fmt.Printf("Login Incorrecto, Intento %d de 3 \n", i)
		}
	}
}

func main() {
	permitido := loginUser()

	if permitido == true {
		authenticatePass()

	}

	go envMensaje(&socketC)
	go recibeReporte(&socketC)
	for {
		if fsalir {
			break
		}

	}
	socketC.Close()

}

func envMensaje(socketC *net.Conn) {

	for {
		// --- Pedir comando ---

		fmt.Println("Digite el comando a enviar:")

		reader := bufio.NewReader(os.Stdin)
		comando, _ := reader.ReadString('\n')
		env := bufio.NewWriter(*socketC)
		env.WriteString(comando + "\n")
		env.Flush()
		if comando == "exit" {
			fsalir = true

		}

		fmt.Println("Client# Comando enviado", comando)
		// ---- Fin de enviar comando al server -----

		// --- Recibir Rta de comando desde el Servidor ---
		rec := bufio.NewReader(*socketC)
		for {
			sResComando, _ := rec.ReadString('\n')
			fmt.Print("Cliente#", sResComando)
			if sResComando == "\n" {
				break
			}
		}
		// --- Fin Recibir Rta de comando desde el Servidor ---
		fmt.Println("-------------------------------------------")
	}
}

func recibeReporte(socketC *net.Conn) {
	for {

		time.Sleep(5 * time.Second)
		rr, _ := bufio.NewReader(*socketC).ReadString('\n')
		fmt.Printf("\rCli# From: %v :::[%s", (*socketC).RemoteAddr, rr)
		fmt.Printf("\nDigite el comando a enviar: ")
	}

}

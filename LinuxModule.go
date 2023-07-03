package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Variables globales
var fsalir bool
var socketS net.Conn

func recibenMensaje(socketS *net.Conn) {
	/*for {
		mr, _ := bufio.NewReader(*socketS).ReadString('\n')
		fmt.Println("Serv# Comando Recibido: ", mr)
		if strings.TrimRight(mr, "\n") == "bye" {
			fsalir = true
		}
	}*/

	for {
		// --- Recibir comando dle cliente ---
		rec := bufio.NewReader(*socketS)
		comando, _ := rec.ReadString('\n')
		fmt.Println("Server# Comando recibido: ...", comando)
		if strings.TrimRight(comando, "\n") == "bye" {
			fsalir = true
		}
		// --- Fin Recibir comando dle cliente ---

		// --- ejecutando el comando ----
		array_datoIn := strings.Fields(comando)
		shell := exec.Command(array_datoIn[0], array_datoIn[1:]...)
		resComando, _ := shell.Output()
		sResComando := string(resComando)
		fmt.Println("Ejecutando comando recibido: ")
		// --- Fin de ejecutando el comando ----

		// ---- Enviar Rta de comando al Cliente -----
		env := bufio.NewWriter(*socketS)
		env.WriteString(sResComando + "\n")
		env.Flush()
		fmt.Println("Server# Respuesta de ejecución enviada!")
		// ---- Fin Enviar Rta de comando al Cliente -----
	}

}

func enviarReporte(socketS *net.Conn) {
	x := 0

	for {
		comando1 := "free -m | awk 'NR==2{print $3}'"
		r1 := exec.Command("bash", "-c", comando1)
		rmem, _ := r1.Output()
		rmem2 := string(rmem)
		rmem2 = strings.TrimRight(rmem2, "\n")
		comando2 := "df -m -t ext4 --output=used /dev/sda2 | tail -n 1 | awk '{print $1}'"
		r2 := exec.Command("bash", "-c", comando2)
		rDD, _ := r2.Output()
		rDD2 := string(rDD)
		rDD2 = strings.TrimRight(rDD2, "\n")
		time.Sleep(5 * time.Second)
		fmt.Println("enviando reporte")
		x++
		mReporte := fmt.Sprintf("******* Reporte de uso #%d -[mem: %s] [DD: %s ]", x, rmem2, rDD2)
		fmt.Println(mReporte)
		eR := bufio.NewWriter(*socketS)
		eR.WriteString(mReporte + "\n")
		eR.Flush()
		fmt.Println("Serv# reporte #", x, " enviado a ", (*socketS).RemoteAddr())
	}

}

func loginRemote() {
	for {
		if fsalir {
			break
		}

		// Recibir el nombre de usuario desde el módulo Windows
		var user string
		err := gob.NewDecoder(socketS).Decode(&user)
		if err != nil {
			fmt.Println("Error al recibir el nombre de usuario:", err)
			break
		}

		// Enviar la lista de usuarios permitidos al módulo Windows
		usuariosPermitidos := obtenerUsuariosPermitidos()
		errd := gob.NewEncoder(socketS).Encode(&usuariosPermitidos)
		if errd != nil {
			fmt.Println("Error al enviar la lista de usuarios permitidos:", err)
			break
		}
		var password string
		for i := 1; i < 4; i++ {

			err = gob.NewDecoder(socketS).Decode(&password)

			if err != nil {
				fmt.Println("Error al recibir la contraseña:", err)
				break
			}

			autenticado := autenticarUsuario(user, password)
			if autenticado == true {
				i = 4
				fmt.Println("Enviando reportes")

			}
			err = gob.NewEncoder(socketS).Encode(&autenticado)
			if err != nil {
				fmt.Println("Error al enviar la autenticación:", err)

			}
			break
		}
		break
	}
}

func obtenerUsuariosPermitidos() []string {
	// Leer el archivo servercommands.conf y obtener la lista de usuarios permitidos
	archivo, err := os.Open("ServerCommands.conf")

	scanner := bufio.NewScanner(archivo)
	for scanner.Scan() {
		linea := scanner.Text()
		if strings.HasPrefix(linea, "Users") {
			// Encontrada la línea con los usuarios permitidos
			usuarios := strings.Split(strings.Split(linea, "=")[1], ",")
			for i := range usuarios {
				usuarios[i] = strings.TrimSpace(usuarios[i])
			}
			return usuarios
		}

	}

	fmt.Println("Error al abrir el archivo de configuración:", err)
	return nil
}

func autenticarUsuario(user string, password string) bool {

	archivo, err := os.Open("users.pw")
	if archivo == nil {
		fmt.Println("Error al abrir el archivo de configuración:", err)
		return false
	}

	scanner := bufio.NewScanner(archivo)
	for scanner.Scan() {
		linea := scanner.Text()
		parts := strings.Split(linea, ":")

		// Verificar si hay al menos 4 partes en el slice
		if len(parts) >= 4 {
			pass := sha256.Sum256([]byte(password))
			passHashIn := fmt.Sprintf("%x", pass)
			passHashDB := parts[3]
			fmt.Println(passHashDB)
			fmt.Println(passHashIn)
			if passHashIn == passHashDB {
				return true
			} else {
				return false
			}

		}
		if scanner.Err() != nil && scanner.Err() != io.EOF {
			fmt.Println("Error al leer el archivo de configuración:", scanner.Err())
		}
	}
	return false
}

func main() {
	fmt.Println("******************++++++********************")
	fmt.Println("[       servidor shell/comandos    ]")
	fmt.Println("============================================")

	tcpAddress, _ := net.ResolveTCPAddr("tcp4", ":1306")
	socketServer, _ := net.ListenTCP("tcp", tcpAddress)
	fmt.Println("Server# Esperando Conexiones ...")
	socketS, _ = socketServer.Accept()
	fmt.Println("Server# se conectó el cliente ...", socketS.RemoteAddr())
	loginRemote()
	go recibenMensaje(&socketS)
	go enviarReporte(&socketS)

	for {
		if fsalir {
			break
		}

	}
}

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

var port = ""
var usuarios []Usuario

type emailStruct struct {
	From    string `json:"from"`
	Nombre  string `json:"nombre"`
	Asunto  string `json:"asunto"`
	Message string `json:"message"`
	Name    string `json:"name"`
}

type AccesoMail struct {
	Identity string `json:"identity"`
	Username string `json:"username"`
	Password string `json:"password"`
	Address  string `json:"address"`
	Host     string `json:"host"`
}
type Usuario struct {
	Name   string     `json:"name"`
	Acceso AccesoMail `json:"acceso"`
}

func BuscarUsuario(nombre string) (Usuario, bool) {
	for _, user := range usuarios {
		if user.Name == nombre {
			return user, true
		}
	}
	return Usuario{}, false
}

func HomeHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Entro home")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	res.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	res.Header().Add("Content-Type", "text/html")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(`<h1>Servidor corriendo en ` + port + `</h1>`))
	// json.NewEncoder(res).Encode("")
}

func MailHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method == "OPTIONS" {
		fmt.Println("Entro options")
		res.Header().Set("Access-Control-Allow-Origin", "*")
		res.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		res.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		res.WriteHeader(http.StatusCreated)
	} else {

		var mail emailStruct
		miMail, err := ioutil.ReadAll(req.Body)
		if err != nil {
			fmt.Fprintf(res, "Error en datos")
		}

		json.Unmarshal(miMail, &mail)
		fmt.Println(mail)
		fmt.Println(BuscarUsuario(mail.Name))
		usuario, encontro := BuscarUsuario(mail.Name)
		if !encontro {
			fmt.Println("Error no existe el usuario ")
			res.Header().Add("Content-Type", "application/json")
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte(`{"message": "No existe el usuario"}`))
		} else {
			var to []string
			to = append(to, getEmailFrom(usuario))
			_, err = sendEmail(usuario, mail.From, mail.Nombre, to, mail.Asunto, []byte(mail.Message))
			if err != nil {
				fmt.Println("Error al enviar Email")
				res.Header().Add("Content-Type", "application/json")
				res.WriteHeader(http.StatusBadRequest)
				res.Write([]byte(`{"message": "Error al enviar Email"}`))
			} else {
				fmt.Println("Exito al enviar Email")
				res.Header().Add("Content-Type", "application/json")
				res.Header().Set("Access-Control-Allow-Origin", "danielahernandez.com.mx localhost")
				res.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
				res.Header().Set("Access-Control-Allow-Headers", "Content-Type")

				res.WriteHeader(http.StatusOK)
				res.Write([]byte(`{"message":"Email enviado"}`))
				// json.NewEncoder(res).Encode("")
			}
		}
	}
}

func sendEmail(usuario Usuario,
	from string, nombre string,
	to []string, asunto string,
	message []byte) (string, error) {
	err := smtp.SendMail(
		getAddressSMTP(usuario),
		plainAuth(usuario), from, to,
		joinMessageStructure(to, from, nombre, asunto, string(message)))
	if err != nil {
		log.Printf("Error!: %s", err)
		return "Error! sending mail", err
	}
	return "email sent successfully", nil
}

func getAddressSMTP(user Usuario) string {
	// return os.Getenv("address")
	return user.Acceso.Address
}

func getEmailFrom(user Usuario) string {
	// return os.Getenv("username")
	return user.Acceso.Username
}

func plainAuth(user Usuario) smtp.Auth {
	return smtp.PlainAuth(
		user.Acceso.Identity,
		user.Acceso.Username,
		user.Acceso.Password,
		user.Acceso.Host)
	// os.Getenv("host"))
}

func joinMessageStructure(
	emailList []string,
	from string,
	nombre string,
	subject string,
	body string) []byte {
	/* func joinMessageStructure(subject string, body string) []byte{ */
	sender := fmt.Sprintf("From: %v\r\n", from)
	concatenate := strings.Join(emailList, ", ")
	toConcatenate := fmt.Sprintf("To: %v\r\n", concatenate)

	subjectMsg := fmt.Sprintf("Subject: %v\r\n", subject)

	bodyMsg2 := fmt.Sprintf("Mi correo es: %v\r\n Mi teléfono es: %v\r\n", from, from)
	bodyMsg := fmt.Sprintf("Hola mi nombre es: %v\r\n%v\r\nMotivo de la consulta: %v",
		nombre, bodyMsg2, body)
	return []byte(sender + toConcatenate + subjectMsg + "\r\n" + bodyMsg)
}

func main() {

	fmt.Println(os.Getenv("ENV"))
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// fmt.Println(os.Getenv("identity"))
	// fmt.Println(os.Getenv("username"))
	// fmt.Println(os.Getenv("password"))
	// fmt.Println(os.Getenv("host"))

	archivo, err := os.Open("usuarios.json")
	if err != nil {
		log.Fatal(err)
	}

	// if os.Getenv("ENV") == "development" {
	// 	port = ":8001"
	// } else {
	// 	port = ":80"
	// }
	port = "localhost:3333"
	fmt.Println(port)

	defer archivo.Close()
	// var usuarios []Usuario
	err = json.NewDecoder(archivo).Decode(&usuarios)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(usuarios)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"https://daniela.jhc-sistemas.com/",
			"https://jose.jhc-sistemas.com/",
			"*",
		},
		AllowedMethods:   []string{"POST", "GET", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})
	var hora = time.Now()
	router := mux.NewRouter()
	rutas := router.PathPrefix("/api").Subrouter()
	rutas.HandleFunc("/", HomeHandler).Methods(http.MethodGet)
	rutas.HandleFunc("/sendmail", MailHandler).Methods(http.MethodPost)
	fmt.Println("Servidor corriendo en " + port + hora.String())
	// c := cors.AllowAll().Handler(rutas)
	// log.Fatal(http.ListenAndServeTLS(port, "server.crt", "server.key", cors.Handler(router)))
	log.Fatal(http.ListenAndServe(port, c.Handler(rutas)))

}

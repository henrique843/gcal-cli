// main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// getClient usa um token de um arquivo ou busca um novo.
func getClient(config *oauth2.Config) *http.Client {
	// O arquivo token.json armazena os tokens de acesso e de atualização do usuário.
	// Ele é criado automaticamente na primeira vez que o fluxo de autorização é completado.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// getTokenFromWeb solicita um token da web.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Acesse esta URL no seu navegador para autorizar o app e cole o código de autorização aqui: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Não foi possível ler o código de autorização: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Não foi possível obter o token a partir do código: %v", err)
	}
	return tok
}

// tokenFromFile recupera um token de um arquivo local.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken salva um token em um arquivo.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Salvando o arquivo de credenciais em: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Não foi possível salvar o token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// --- Comandos do CLI ---

// rootCmd representa o comando base quando chamado sem subcomandos
var rootCmd = &cobra.Command{
	Use:   "gcal",
	Short: "Um CLI para interagir com o Google Calendar.",
	Long: `gcal é uma ferramenta de linha de comando para listar e criar
eventos no seu Google Calendar.`,
}

// listCmd para listar eventos
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista os próximos 10 eventos.",
	Run: func(cmd *cobra.Command, args []string) {
		b, err := ioutil.ReadFile("credentials.json")
		if err != nil {
			log.Fatalf("Não foi possível ler o arquivo de credenciais: %v", err)
		}

		// Se você modificar esses escopos, delete o arquivo token.json anterior.
		config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
		if err != nil {
			log.Fatalf("Não foi possível processar o arquivo de credenciais: %v", err)
		}
		client := getClient(config)

		srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
		if err != nil {
			log.Fatalf("Não foi possível obter o cliente do Calendar: %v", err)
		}

		t := time.Now().Format(time.RFC3339)
		events, err := srv.Events.List("primary").
			ShowDeleted(false).
			SingleEvents(true).
			TimeMin(t).
			MaxResults(10).
			OrderBy("startTime").Do()
		if err != nil {
			log.Fatalf("Não foi possível obter os próximos eventos: %v", err)
		}

		fmt.Println("Próximos 10 eventos:")
		if len(events.Items) == 0 {
			fmt.Println("Nenhum evento próximo encontrado.")
		} else {
			for _, item := range events.Items {
				date := item.Start.DateTime
				if date == "" { // Evento de dia inteiro
					date = item.Start.Date
				}
				fmt.Printf("• %v (%v)\n", item.Summary, date)
			}
		}
	},
}

// addCmd para criar um novo evento
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Adiciona um novo evento ao calendário.",
	Run: func(cmd *cobra.Command, args []string) {
		// Obter os dados do evento a partir das flags
		title, _ := cmd.Flags().GetString("title")
		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")
		description, _ := cmd.Flags().GetString("desc")

		if title == "" || start == "" || end == "" {
			log.Fatalf("Flags --title, --start e --end são obrigatórias.")
		}

		b, err := ioutil.ReadFile("credentials.json")
		if err != nil {
			log.Fatalf("Não foi possível ler o arquivo de credenciais: %v", err)
		}

		// O escopo agora precisa de permissão de escrita!
		config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
		if err != nil {
			log.Fatalf("Não foi possível processar o arquivo de credenciais: %v", err)
		}
		client := getClient(config)

		srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
		if err != nil {
			log.Fatalf("Não foi possível obter o cliente do Calendar: %v", err)
		}

		event := &calendar.Event{
			Summary:     title,
			Description: description,
			Start: &calendar.EventDateTime{
				DateTime: start,
				TimeZone: "America/Sao_Paulo", // Mude para seu fuso horário
			},
			End: &calendar.EventDateTime{
				DateTime: end,
				TimeZone: "America/Sao_Paulo",
			},
		}

		calendarId := "primary"
		event, err = srv.Events.Insert(calendarId, event).Do()
		if err != nil {
			log.Fatalf("Não foi possível criar o evento: %v", err)
		}

		fmt.Printf("Evento criado com sucesso! Link: %s\n", event.HtmlLink)
	},
}

func main() {
	// Adiciona as flags para o comando 'add'
	addCmd.Flags().String("title", "", "Título do evento")
	addCmd.Flags().String("start", "", "Horário de início (formato: 2025-09-08T10:00:00-03:00)")
	addCmd.Flags().String("end", "", "Horário de término (formato: 2025-09-08T11:00:00-03:00)")
	addCmd.Flags().String("desc", "", "Descrição do evento")

	// Adiciona os subcomandos ao comando raiz
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)

	// Executa o comando raiz
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

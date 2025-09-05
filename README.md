# Go Google Calendar CLI

Este é um gerenciador de eventos para o Google Calendar feito em Golang. Ele permite a manipulação da agenda via linha de comando.

## Funcionalidades

- Listar os próximos 10 eventos.
- Adicionar novos eventos ao calendário.

## Setup

Para rodar este projeto, você precisará de credenciais da API do Google Calendar.

1. Siga os passos da [documentação oficial](https://developers.google.com/calendar/api/quickstart/go) para criar seu projeto no Google Cloud e obter o arquivo `credentials.json`.
2. Coloque o arquivo `credentials.json` na raiz deste projeto.
3. Rode `go run main.go list` para autorizar a aplicação.

## Uso

**Listar eventos:**
```bash
go run main.go list
```

**Adicionar um evento:**
```bash
go run main.go add --title "Meu Evento" --start "YYYY-MM-DDTHH:MM:SS-03:00" --end "YYYY-MM-DDTHH:MM:SS-03:00"
```

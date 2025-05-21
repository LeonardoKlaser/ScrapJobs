# ScrapJobs

ScrapJobs é um projeto desenvolvido em Go para realizar o scraping de vagas de emprego, armazenando as informações em um banco de dados PostgreSQL e, caso encontre novas vagas, notificar por e-mail via AWS SES.

## Funcionalidades

- **Web Scraping de Vagas:** Coleta vagas de emprego (com foco em "developer" e "software") do site [jobs.sap.com](https://jobs.sap.com/search/?q=&locationsearch=S%C3%A3o+Leopoldo&location=S%C3%A3o+Leopoldo&scrollToTable=true), extraindo dados como título, localização, empresa, link da vaga e ID da requisição.
- **Banco de Dados:** Armazena as vagas encontradas em um banco PostgreSQL, prevenindo duplicatas através do ID da requisição.
- **Envio de Notificações por E-mail:** Ao identificar novas vagas, envia e-mail de notificação usando AWS SES.
- **API REST:** Disponibiliza um endpoint HTTP `/scrape` para acionar o scraper manualmente e inserir novas vagas.

## Tecnologias Utilizadas

- [Go](https://golang.org/)
- [Gin](https://github.com/gin-gonic/gin) (framework web)
- [Colly](https://github.com/gocolly/colly) (web scraping)
- [PostgreSQL](https://www.postgresql.org/) (banco de dados relacional)
- [AWS SES](https://aws.amazon.com/ses/) (envio de e-mails)

## Como Executar

### Pré-requisitos

- Docker (opcional, para facilitar a execução)
- Variáveis/configurações da AWS para envio de e-mail via SES
- Banco PostgreSQL disponível

### Usando Docker

1. **Clone o repositório:**
   ```bash
   git clone https://github.com/LeonardoKlaser/ScrapJobs.git
   cd ScrapJobs
   ```

2. **Build e execução (ajuste variáveis de ambiente se necessário):**
   ```bash
   docker-compose up -d --build .
   ```

### Executando Localmente

1. Instale as dependências do Go.
2. Ajuste as configurações de banco e AWS conforme necessário.
3. Execute:
   ```bash
   go run scrapper.go
   ```

## Estrutura Básica

- `scrapper.go` — Inicialização da aplicação e rotas HTTP.
- `scrapper/` — Lógica de scraping web.
- `repository/` — Acesso ao banco de dados.
- `infra/db/` — Conexão com o banco de dados.
- `infra/ses/` — Integração com AWS SES para envio de e-mails.
- `usecase/` — Regras de negócio (armazenamento e notificação).
- `controller/` — Camada de controle da API.

## Exemplo de Uso da API

- **GET** `/scrape`
  - Aciona manualmente a coleta de vagas e retorna as novas inseridas.

## Observações

- As vagas são filtradas por palavras-chave no título (“developer” ou “software”) e localizadas em São Leopoldo.
- É necessário configurar corretamente as credenciais AWS e permissões SES para envio de e-mails.
- O projeto pode ser expandido para rastrear outros sites e/ou cidades alterando a lógica do scraper.


Desenvolvido por [Leonardo Klaser](https://github.com/LeonardoKlaser)

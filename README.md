# ScrapJobs

ScrapJobs é um projeto desenvolvido em Go para realizar o scraping de vagas de emprego, armazenando as informações em um banco de dados PostgreSQL. Quando novas vagas são encontradas, o sistema notifica os usuários por e-mail, fornecendo um link para a vaga e uma análise de compatibilidade entre o currículo do usuário e os requisitos da vaga, incluindo dicas de melhorias e pontos fortes.

## Funcionamento do Sistema

O sistema opera em uma instância EC2 da AWS e é acionado automaticamente em dias úteis, de hora em hora, durante o horário comercial (aproximadamente das 08:00 às 18:00). A cada hora, a rotina de scraping é executada para buscar e analisar novas vagas.

## Funcionalidades Principais

- **Web Scraping de Vagas:** Coleta vagas de emprego (com foco em "developer" e "software") em sites de grandes empresas, extraindo dados como título, localização, empresa, link da vaga e ID da requisição.
- **Análise de Compatibilidade (Match) com Currículo:** Ao encontrar uma nova vaga, o sistema compara os requisitos da vaga com o currículo cadastrado pelo usuário.
- **Notificações Inteligentes por E-mail:** Usuários são alertados via AWS SES sobre novas vagas. O e-mail inclui:
    - Link direto para a vaga.
    - Percentual de compatibilidade entre o currículo do usuário e a vaga.
    - Destaque dos pontos fortes do currículo em relação à vaga.
    - Sugestões de melhorias no currículo para aumentar a compatibilidade.
- **Banco de Dados:** Armazena as vagas encontradas e informações de usuários/currículos em um banco PostgreSQL, prevenindo duplicatas de vagas através do ID da requisição.
- **API REST:** Disponibiliza um endpoint HTTP `/scrape` para acionar o scraper manualmente e inserir novas vagas (utilizado principalmente para desenvolvimento e testes).

## Tecnologias Utilizadas

- [Go](https://golang.org/)
- [Gin](https://github.com/gin-gonic/gin) (framework web)
- [Colly](https://github.com/gocolly/colly) (web scraping)
- [PostgreSQL](https://www.postgresql.org/) (banco de dados relacional)
- [AWS EC2](https://aws.amazon.com/ec2/) (hospedagem da aplicação)
- [AWS SES](https://aws.amazon.com/ses/) (envio de e-mails)
- [Docker](https://www.docker.com/) (para facilitar a execução e deploy)
- *[Mencionar aqui bibliotecas de IA/NLP, se aplicável, para a análise de currículo vs vaga]*

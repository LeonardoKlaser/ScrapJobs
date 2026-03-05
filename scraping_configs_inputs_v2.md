# Configurações de Scraping v2 — Inputs para o Banco de Dados

Data: 2026-03-04

## Padrões de ATS

### Greenhouse (API)

Empresas: Nubank, XP Inc., QuintoAndar, Wildlife Studios, PicPay, C6 Bank

```
Endpoint: https://boards-api.greenhouse.io/v1/boards/{SLUG}/jobs?content=true
Método: GET
JSONDataMappings: { "jobs_array_path": "jobs", "title_path": "title", "link_path": "absolute_url", "location_path": "location.name", "description_path": "content", "requisition_id_path": "id" }
```

### Gupy (HTML)

Empresas: Itaú, Ambev, Vivo, Creditas, Stone Tech, Grupo Boticário, Embraer, Sicredi

```
URL: https://{SLUG}.gupy.io/
Tipo: HTML (CSS Selectors via Colly)
⚠️ PROIBIDO usar _next/data/{buildId} — o buildId muda a cada deploy do Gupy.
```

### Eightfold AI (API)

Empresas: Mercado Livre, Vale

```
Endpoint Mercado Livre: https://mercadolibre.eightfold.ai/api/pcsx/search?domain=mercadolibre.com&start=0&num=100
Endpoint Vale: https://vale.eightfold.ai/api/apply/v2/jobs?domain=vale.com&start=0&num=100
```

---

## Configurações das Empresas

### 1. Nubank (Greenhouse)

```json
{
  "SiteName": "Nubank Careers",
  "BaseURL": "https://job-boards.greenhouse.io/nubank",
  "IsActive": true,
  "ScrapingType": "API",
  "JobListItemSelector": null,
  "TitleSelector": null,
  "LinkSelector": null,
  "LinkAttribute": null,
  "LocationSelector": null,
  "NextPageSelector": null,
  "JobDescriptionSelector": null,
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": "https://boards-api.greenhouse.io/v1/boards/nubank/jobs?content=true",
  "APIMethod": "GET",
  "APIHeadersJSON": "{\"Content-Type\": \"application/json\"}",
  "APIPayloadTemplate": null,
  "JSONDataMappings": "{ \"jobs_array_path\": \"jobs\", \"title_path\": \"title\", \"link_path\": \"absolute_url\", \"location_path\": \"location.name\", \"description_path\": \"content\", \"requisition_id_path\": \"id\" }"
}
```

---

### 2. Itaú Unibanco (Gupy)

> **Mudança v1→v2:** Trocado de API (`_next/data`) para HTML (CSS Selectors).

```json
{
  "SiteName": "Itaú Unibanco Carreiras",
  "BaseURL": "https://vemproitau.gupy.io/",
  "IsActive": true,
  "ScrapingType": "HTML",
  "JobListItemSelector": "ul[aria-label='Lista de vagas'] li, ul[data-testid='job-list'] li",
  "TitleSelector": "h2, h3",
  "LinkSelector": "a",
  "LinkAttribute": "href",
  "LocationSelector": "span[data-testid='job-location'], div[class*='location']",
  "NextPageSelector": "button[aria-label='Próxima página'], a[aria-label='Next']",
  "JobDescriptionSelector": "div[data-testid='job-description']",
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": null,
  "APIMethod": null,
  "APIHeadersJSON": null,
  "APIPayloadTemplate": null,
  "JSONDataMappings": null
}
```

---

### 3. Ambev (Gupy)

> **Mudança v1→v2:** Trocado de API (`_next/data`) para HTML (CSS Selectors).

```json
{
  "SiteName": "Ambev Carreiras",
  "BaseURL": "https://ambev.gupy.io/",
  "IsActive": true,
  "ScrapingType": "HTML",
  "JobListItemSelector": "ul[aria-label='Lista de vagas'] li, ul[data-testid='job-list'] li",
  "TitleSelector": "h2, h3",
  "LinkSelector": "a",
  "LinkAttribute": "href",
  "LocationSelector": "span[data-testid='job-location'], div[class*='location']",
  "NextPageSelector": "button[aria-label='Próxima página'], a[aria-label='Next']",
  "JobDescriptionSelector": "div[data-testid='job-description']",
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": null,
  "APIMethod": null,
  "APIHeadersJSON": null,
  "APIPayloadTemplate": null,
  "JSONDataMappings": null
}
```

---

### 4. Mercado Livre (Eightfold AI)

```json
{
  "SiteName": "Mercado Livre Carreiras",
  "BaseURL": "https://mercadolibre.eightfold.ai/careers",
  "IsActive": true,
  "ScrapingType": "API",
  "JobListItemSelector": null,
  "TitleSelector": null,
  "LinkSelector": null,
  "LinkAttribute": null,
  "LocationSelector": null,
  "NextPageSelector": null,
  "JobDescriptionSelector": null,
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": "https://mercadolibre.eightfold.ai/api/pcsx/search?domain=mercadolibre.com&start=0&num=100",
  "APIMethod": "GET",
  "APIHeadersJSON": "{\"Content-Type\": \"application/json\"}",
  "APIPayloadTemplate": null,
  "JSONDataMappings": "{ \"jobs_array_path\": \"data.positions\", \"title_path\": \"name\", \"link_path\": \"positionUrl\", \"location_path\": \"locations.0\", \"description_path\": \"\", \"requisition_id_path\": \"id\" }"
}
```

---

### 5. XP Inc. (Greenhouse)

```json
{
  "SiteName": "XP Inc. Carreiras",
  "BaseURL": "https://job-boards.greenhouse.io/xpinc",
  "IsActive": true,
  "ScrapingType": "API",
  "JobListItemSelector": null,
  "TitleSelector": null,
  "LinkSelector": null,
  "LinkAttribute": null,
  "LocationSelector": null,
  "NextPageSelector": null,
  "JobDescriptionSelector": null,
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": "https://boards-api.greenhouse.io/v1/boards/xpinc/jobs?content=true",
  "APIMethod": "GET",
  "APIHeadersJSON": "{\"Content-Type\": \"application/json\"}",
  "APIPayloadTemplate": null,
  "JSONDataMappings": "{ \"jobs_array_path\": \"jobs\", \"title_path\": \"title\", \"link_path\": \"absolute_url\", \"location_path\": \"location.name\", \"description_path\": \"content\", \"requisition_id_path\": \"id\" }"
}
```

---

### 6. Vivo (Gupy)

> **Mudança v1→v2:** Trocado de API (`_next/data`) para HTML (CSS Selectors).

```json
{
  "SiteName": "Vivo Carreiras",
  "BaseURL": "https://vivo.gupy.io/",
  "IsActive": true,
  "ScrapingType": "HTML",
  "JobListItemSelector": "ul[aria-label='Lista de vagas'] li, ul[data-testid='job-list'] li",
  "TitleSelector": "h2, h3",
  "LinkSelector": "a",
  "LinkAttribute": "href",
  "LocationSelector": "span[data-testid='job-location'], div[class*='location']",
  "NextPageSelector": "button[aria-label='Próxima página'], a[aria-label='Next']",
  "JobDescriptionSelector": "div[data-testid='job-description']",
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": null,
  "APIMethod": null,
  "APIHeadersJSON": null,
  "APIPayloadTemplate": null,
  "JSONDataMappings": null
}
```

---

### 7. QuintoAndar (Greenhouse)

```json
{
  "SiteName": "QuintoAndar Carreiras",
  "BaseURL": "https://job-boards.greenhouse.io/quintoandar",
  "IsActive": true,
  "ScrapingType": "API",
  "JobListItemSelector": null,
  "TitleSelector": null,
  "LinkSelector": null,
  "LinkAttribute": null,
  "LocationSelector": null,
  "NextPageSelector": null,
  "JobDescriptionSelector": null,
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": "https://boards-api.greenhouse.io/v1/boards/quintoandar/jobs?content=true",
  "APIMethod": "GET",
  "APIHeadersJSON": "{\"Content-Type\": \"application/json\"}",
  "APIPayloadTemplate": null,
  "JSONDataMappings": "{ \"jobs_array_path\": \"jobs\", \"title_path\": \"title\", \"link_path\": \"absolute_url\", \"location_path\": \"location.name\", \"description_path\": \"content\", \"requisition_id_path\": \"id\" }"
}
```

---

### 8. Creditas (Gupy)

```json
{
  "SiteName": "Creditas Carreiras",
  "BaseURL": "https://creditas.gupy.io/",
  "IsActive": true,
  "ScrapingType": "HTML",
  "JobListItemSelector": "ul[aria-label='Lista de vagas'] li, ul[data-testid='job-list'] li",
  "TitleSelector": "h2, h3",
  "LinkSelector": "a",
  "LinkAttribute": "href",
  "LocationSelector": "span[data-testid='job-location'], div[class*='location']",
  "NextPageSelector": "button[aria-label='Próxima página'], a[aria-label='Next']",
  "JobDescriptionSelector": "div[data-testid='job-description']",
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": null,
  "APIMethod": null,
  "APIHeadersJSON": null,
  "APIPayloadTemplate": null,
  "JSONDataMappings": null
}
```

---

### 9. Stone (Gupy — Stone Tech)

> **Nota:** Stone tem portal principal customizado em `jornada.stone.com.br` (SPA, requer HEADLESS). Esta config usa o subdomain Gupy da divisão de tecnologia. Também existe `stonebanking.gupy.io` para Stone Banking.

```json
{
  "SiteName": "Stone Tech Carreiras",
  "BaseURL": "https://stech.gupy.io/",
  "IsActive": true,
  "ScrapingType": "HTML",
  "JobListItemSelector": "ul[aria-label='Lista de vagas'] li, ul[data-testid='job-list'] li",
  "TitleSelector": "h2, h3",
  "LinkSelector": "a",
  "LinkAttribute": "href",
  "LocationSelector": "span[data-testid='job-location'], div[class*='location']",
  "NextPageSelector": "button[aria-label='Próxima página'], a[aria-label='Next']",
  "JobDescriptionSelector": "div[data-testid='job-description']",
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": null,
  "APIMethod": null,
  "APIHeadersJSON": null,
  "APIPayloadTemplate": null,
  "JSONDataMappings": null
}
```

---

### 10. Magazine Luiza (InHire)

> **⚠️ HEADLESS necessário:** InHire é uma SPA JavaScript pura — sem HTML server-side. Seletores abaixo são estimativas baseadas em padrões comuns do InHire. Necessitam validação via inspeção do DOM renderizado no browser (DevTools → Elements).
>
> Alternativa: `carreiras.magazineluiza.com.br` (WordPress) também carrega vagas via JS.

```json
{
  "SiteName": "Magazine Luiza Carreiras",
  "BaseURL": "https://magazineluiza.inhire.app/vagas",
  "IsActive": true,
  "ScrapingType": "HEADLESS",
  "JobListItemSelector": "div[class*='job-item'], div[class*='vacancy-item'], a[class*='job-card']",
  "TitleSelector": "h3, h2, span[class*='title']",
  "LinkSelector": "a",
  "LinkAttribute": "href",
  "LocationSelector": "span[class*='location'], p[class*='location']",
  "NextPageSelector": "button[class*='next'], a[class*='next']",
  "JobDescriptionSelector": "div[class*='description'], div[class*='content']",
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": null,
  "APIMethod": null,
  "APIHeadersJSON": null,
  "APIPayloadTemplate": null,
  "JSONDataMappings": null
}
```

---

### 11. Grupo Boticário (Gupy)

```json
{
  "SiteName": "Grupo Boticário Carreiras",
  "BaseURL": "https://grupoboticario.gupy.io/",
  "IsActive": true,
  "ScrapingType": "HTML",
  "JobListItemSelector": "ul[aria-label='Lista de vagas'] li, ul[data-testid='job-list'] li",
  "TitleSelector": "h2, h3",
  "LinkSelector": "a",
  "LinkAttribute": "href",
  "LocationSelector": "span[data-testid='job-location'], div[class*='location']",
  "NextPageSelector": "button[aria-label='Próxima página'], a[aria-label='Next']",
  "JobDescriptionSelector": "div[data-testid='job-description']",
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": null,
  "APIMethod": null,
  "APIHeadersJSON": null,
  "APIPayloadTemplate": null,
  "JSONDataMappings": null
}
```

---

### 12. TOTVS (Trakstar Hire)

> **Nota:** TOTVS tem ATS próprio em `atracaodetalentos.totvs.app/vempratotvs` mas também publica vagas no Trakstar Hire com HTML server-side. Seletores baseados em classes semânticas do Trakstar (`js-*`). Validar via DevTools.

```json
{
  "SiteName": "TOTVS Carreiras",
  "BaseURL": "https://totvs.hire.trakstar.com/",
  "IsActive": true,
  "ScrapingType": "HTML",
  "JobListItemSelector": ".js-careers-page-job-list-item",
  "TitleSelector": ".js-job-list-opening-name, h3",
  "LinkSelector": "a",
  "LinkAttribute": "href",
  "LocationSelector": ".js-job-list-city",
  "NextPageSelector": null,
  "JobDescriptionSelector": ".js-job-description, .job-description",
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": null,
  "APIMethod": null,
  "APIHeadersJSON": null,
  "APIPayloadTemplate": null,
  "JSONDataMappings": null
}
```

---

### 13. Vale (Eightfold AI)

> **Nota:** Endpoint diferente do Mercado Livre (`/api/apply/v2/jobs` vs `/api/pcsx/search`). Retorna `positions` array direto (sem wrapper `data`). `canonicalPositionUrl` é URL absoluta. Suporta paginação via `start` + `num`.

```json
{
  "SiteName": "Vale Carreiras",
  "BaseURL": "https://vale.eightfold.ai/careers",
  "IsActive": true,
  "ScrapingType": "API",
  "JobListItemSelector": null,
  "TitleSelector": null,
  "LinkSelector": null,
  "LinkAttribute": null,
  "LocationSelector": null,
  "NextPageSelector": null,
  "JobDescriptionSelector": null,
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": "https://vale.eightfold.ai/api/apply/v2/jobs?domain=vale.com&start=0&num=100",
  "APIMethod": "GET",
  "APIHeadersJSON": "{\"Content-Type\": \"application/json\"}",
  "APIPayloadTemplate": null,
  "JSONDataMappings": "{ \"jobs_array_path\": \"positions\", \"title_path\": \"name\", \"link_path\": \"canonicalPositionUrl\", \"location_path\": \"location\", \"description_path\": \"job_description\", \"requisition_id_path\": \"id\" }"
}
```

---

### 14. Natura (Workday)

> **⚠️ Requer melhoria no API Scraper:** O endpoint Workday usa POST com body JSON. O `APIScrapper` atual passa `nil` como body (linha 41 de `api_scrapper.go`). Para funcionar, o scraper precisa ser atualizado para enviar `APIPayloadTemplate` como body quando `APIMethod` = POST.
>
> Alternativa: usar HEADLESS para `natura.wd501.myworkdayjobs.com/NaturaCarreiras`.

```json
{
  "SiteName": "Natura Carreiras",
  "BaseURL": "https://natura.wd501.myworkdayjobs.com/NaturaCarreiras",
  "IsActive": false,
  "ScrapingType": "API",
  "JobListItemSelector": null,
  "TitleSelector": null,
  "LinkSelector": null,
  "LinkAttribute": null,
  "LocationSelector": null,
  "NextPageSelector": null,
  "JobDescriptionSelector": null,
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": "https://natura.wd501.myworkdayjobs.com/wday/cxs/natura/NaturaCarreiras/jobs",
  "APIMethod": "POST",
  "APIHeadersJSON": "{\"Content-Type\": \"application/json\"}",
  "APIPayloadTemplate": "{\"limit\": 100, \"offset\": 0}",
  "JSONDataMappings": "{ \"jobs_array_path\": \"jobPostings\", \"title_path\": \"title\", \"link_path\": \"externalPath\", \"location_path\": \"locationsText\", \"description_path\": \"\", \"requisition_id_path\": \"bulletFields.0\" }"
}
```

---

### 15. Embraer (Gupy)

```json
{
  "SiteName": "Embraer Carreiras",
  "BaseURL": "https://embraer.gupy.io/",
  "IsActive": true,
  "ScrapingType": "HTML",
  "JobListItemSelector": "ul[aria-label='Lista de vagas'] li, ul[data-testid='job-list'] li",
  "TitleSelector": "h2, h3",
  "LinkSelector": "a",
  "LinkAttribute": "href",
  "LocationSelector": "span[data-testid='job-location'], div[class*='location']",
  "NextPageSelector": "button[aria-label='Próxima página'], a[aria-label='Next']",
  "JobDescriptionSelector": "div[data-testid='job-description']",
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": null,
  "APIMethod": null,
  "APIHeadersJSON": null,
  "APIPayloadTemplate": null,
  "JSONDataMappings": null
}
```

---

### 16. Cielo (InHire)

> **⚠️ HEADLESS necessário:** InHire é uma SPA JavaScript pura. Seletores abaixo são estimativas. Necessitam validação via inspeção do DOM renderizado no browser (DevTools → Elements).

```json
{
  "SiteName": "Cielo Carreiras",
  "BaseURL": "https://cielo.inhire.app/vagas",
  "IsActive": true,
  "ScrapingType": "HEADLESS",
  "JobListItemSelector": "div[class*='job-item'], div[class*='vacancy-item'], a[class*='job-card']",
  "TitleSelector": "h3, h2, span[class*='title']",
  "LinkSelector": "a",
  "LinkAttribute": "href",
  "LocationSelector": "span[class*='location'], p[class*='location']",
  "NextPageSelector": "button[class*='next'], a[class*='next']",
  "JobDescriptionSelector": "div[class*='description'], div[class*='content']",
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": null,
  "APIMethod": null,
  "APIHeadersJSON": null,
  "APIPayloadTemplate": null,
  "JSONDataMappings": null
}
```

---

### 17. Wildlife Studios (Greenhouse)

```json
{
  "SiteName": "Wildlife Studios Carreiras",
  "BaseURL": "https://job-boards.greenhouse.io/wildlifestudios",
  "IsActive": true,
  "ScrapingType": "API",
  "JobListItemSelector": null,
  "TitleSelector": null,
  "LinkSelector": null,
  "LinkAttribute": null,
  "LocationSelector": null,
  "NextPageSelector": null,
  "JobDescriptionSelector": null,
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": "https://boards-api.greenhouse.io/v1/boards/wildlifestudios/jobs?content=true",
  "APIMethod": "GET",
  "APIHeadersJSON": "{\"Content-Type\": \"application/json\"}",
  "APIPayloadTemplate": null,
  "JSONDataMappings": "{ \"jobs_array_path\": \"jobs\", \"title_path\": \"title\", \"link_path\": \"absolute_url\", \"location_path\": \"location.name\", \"description_path\": \"content\", \"requisition_id_path\": \"id\" }"
}
```

---

### 18. PicPay (Greenhouse)

```json
{
  "SiteName": "PicPay Carreiras",
  "BaseURL": "https://job-boards.greenhouse.io/picpay",
  "IsActive": true,
  "ScrapingType": "API",
  "JobListItemSelector": null,
  "TitleSelector": null,
  "LinkSelector": null,
  "LinkAttribute": null,
  "LocationSelector": null,
  "NextPageSelector": null,
  "JobDescriptionSelector": null,
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": "https://boards-api.greenhouse.io/v1/boards/picpay/jobs?content=true",
  "APIMethod": "GET",
  "APIHeadersJSON": "{\"Content-Type\": \"application/json\"}",
  "APIPayloadTemplate": null,
  "JSONDataMappings": "{ \"jobs_array_path\": \"jobs\", \"title_path\": \"title\", \"link_path\": \"absolute_url\", \"location_path\": \"location.name\", \"description_path\": \"content\", \"requisition_id_path\": \"id\" }"
}
```

---

### 19. C6 Bank (Greenhouse)

```json
{
  "SiteName": "C6 Bank Carreiras",
  "BaseURL": "https://job-boards.greenhouse.io/c6bank",
  "IsActive": true,
  "ScrapingType": "API",
  "JobListItemSelector": null,
  "TitleSelector": null,
  "LinkSelector": null,
  "LinkAttribute": null,
  "LocationSelector": null,
  "NextPageSelector": null,
  "JobDescriptionSelector": null,
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": "https://boards-api.greenhouse.io/v1/boards/c6bank/jobs?content=true",
  "APIMethod": "GET",
  "APIHeadersJSON": "{\"Content-Type\": \"application/json\"}",
  "APIPayloadTemplate": null,
  "JSONDataMappings": "{ \"jobs_array_path\": \"jobs\", \"title_path\": \"title\", \"link_path\": \"absolute_url\", \"location_path\": \"location.name\", \"description_path\": \"content\", \"requisition_id_path\": \"id\" }"
}
```

---

### 20. Sicredi (Gupy)

```json
{
  "SiteName": "Sicredi Carreiras",
  "BaseURL": "https://sicredi.gupy.io/",
  "IsActive": true,
  "ScrapingType": "HTML",
  "JobListItemSelector": "ul[aria-label='Lista de vagas'] li, ul[data-testid='job-list'] li",
  "TitleSelector": "h2, h3",
  "LinkSelector": "a",
  "LinkAttribute": "href",
  "LocationSelector": "span[data-testid='job-location'], div[class*='location']",
  "NextPageSelector": "button[aria-label='Próxima página'], a[aria-label='Next']",
  "JobDescriptionSelector": "div[data-testid='job-description']",
  "JobRequisitionIdSelector": null,
  "APIEndpointTemplate": null,
  "APIMethod": null,
  "APIHeadersJSON": null,
  "APIPayloadTemplate": null,
  "JSONDataMappings": null
}
```

---

## Resumo

| # | Empresa | ATS | Tipo | Slug/URL | Status |
|---|---------|-----|------|----------|--------|
| 1 | Nubank | Greenhouse | API | `nubank` | Pronto |
| 2 | Itaú Unibanco | Gupy | HTML | `vemproitau` | Pronto (v1→v2) |
| 3 | Ambev | Gupy | HTML | `ambev` | Pronto (v1→v2) |
| 4 | Mercado Livre | Eightfold AI | API | `mercadolibre` | Pronto |
| 5 | XP Inc. | Greenhouse | API | `xpinc` | Pronto |
| 6 | Vivo | Gupy | HTML | `vivo` | Pronto (v1→v2) |
| 7 | QuintoAndar | Greenhouse | API | `quintoandar` | Pronto |
| 8 | Creditas | Gupy | HTML | `creditas` | Pronto |
| 9 | Stone Tech | Gupy | HTML | `stech` | Pronto |
| 10 | Magazine Luiza | InHire | HEADLESS | `magazineluiza.inhire.app` | ⚠️ Validar seletores |
| 11 | Grupo Boticário | Gupy | HTML | `grupoboticario` | Pronto |
| 12 | TOTVS | Trakstar Hire | HTML | `totvs.hire.trakstar.com` | ⚠️ Validar seletores |
| 13 | Vale | Eightfold AI | API | `vale` | Pronto |
| 14 | Natura | Workday | API (POST) | `natura.wd501` | ⚠️ Requer POST body no scraper |
| 15 | Embraer | Gupy | HTML | `embraer` | Pronto |
| 16 | Cielo | InHire | HEADLESS | `cielo.inhire.app` | ⚠️ Validar seletores |
| 17 | Wildlife Studios | Greenhouse | API | `wildlifestudios` | Pronto |
| 18 | PicPay | Greenhouse | API | `picpay` | Pronto |
| 19 | C6 Bank | Greenhouse | API | `c6bank` | Pronto |
| 20 | Sicredi | Gupy | HTML | `sicredi` | Pronto |

### Contagem por ATS

| ATS | Tipo | Qtd | Empresas |
|-----|------|-----|----------|
| Greenhouse | API | 6 | Nubank, XP, QuintoAndar, Wildlife Studios, PicPay, C6 Bank |
| Gupy | HTML | 8 | Itaú, Ambev, Vivo, Creditas, Stone Tech, Grupo Boticário, Embraer, Sicredi |
| Eightfold AI | API | 2 | Mercado Livre, Vale |
| InHire | HEADLESS | 2 | Magazine Luiza, Cielo |
| Trakstar Hire | HTML | 1 | TOTVS |
| Workday | API (POST) | 1 | Natura |

### Itens de Ação

1. **Validar seletores HEADLESS** — Magazine Luiza e Cielo (InHire SPA): abrir DevTools no browser, renderizar a página, inspecionar o DOM para confirmar/ajustar os seletores CSS.
2. **Validar seletores Trakstar** — TOTVS: confirmar classes `js-*` no DOM renderizado.
3. **Implementar POST body no API Scraper** — Para Natura (Workday): alterar `api_scrapper.go` linha 41 para enviar `APIPayloadTemplate` como body quando `APIMethod` = "POST". Marcar Natura como `IsActive: true` após implementação.
4. **Validar seletores Gupy** — Testar ao menos um site Gupy (ex: `creditas.gupy.io`) para confirmar que os seletores `data-testid` e `aria-label` funcionam com Colly (HTML server-side). Gupy renderiza ~10 vagas no HTML inicial.

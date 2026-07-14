# Banco de dados LicitaHub

O schema principal esta em:

```text
database/schema.sql
```

Banco local usado no desenvolvimento:

```text
licitahub_dev
```

## Stack

- PostgreSQL 17 no ambiente local.
- Extensao `pgcrypto`.
- Migracoes seguras tambem sao aplicadas pelo backend ao iniciar.

## Grupos de tabelas

### Acesso e administracao

- `companies`
- `access_profiles`
- `users`
- `company_invitations`
- `auth_sessions`
- `password_reset_tokens`
- `company_reviews`

### Empresa

- `company_profiles`
- `technical_areas`
- `company_technical_areas`
- `media_files`

### Radar LicitaHub

- `news_categories`
- `news`

### Comunidade

- `post_categories`
- `posts`
- `post_images`
- `post_likes`
- `post_comments`
- `post_favorites`

### Editais

- `tenders`
- `tender_files`
- `tender_requirement_types`
- `tender_requirements`
- `tender_interests`
- `tender_interest_requirements`

### Match e consorcio

- `partnership_ads`
- `partner_evaluations`
- `matches`
- `match_contacts`
- `consortium_intentions`
- `consortium_members`
- `consortium_applications`
- `chat_threads`
- `chat_messages`
- `chat_thread_reads`

### Central de Montagem

- `bid_assembly_templates`
- `bid_assembly_template_stages`
- `bid_assembly_template_tasks`
- `bid_assemblies`
- `bid_assembly_participants`
- `bid_assembly_stages`
- `bid_assembly_tasks`
- `bid_assembly_task_comments`
- `bid_assembly_task_evidences`
- `bid_assembly_deadline_alerts`
- `bid_assembly_activity_logs`

### Notificacoes e auditoria

- `notifications`
- `audit_logs`

A tabela `audit_logs` existe como base tecnica, mas a auditoria funcional completa fica para fase posterior.

## Regras importantes

- Nome fantasia da empresa e unico.
- CNPJ da empresa e unico.
- Email de usuario pode repetir conforme regra definida no MVP.
- Usuarios removidos nao aparecem como ativos.
- Noticias vencidas nao aparecem no Radar publico.
- Anuncios consorciados podem ser fechados para sair da vitrine publica.
- Membros de consorcio podem ser marcados como retirados, preservando data e usuario responsavel pela desistência.
- Chat e mensagens ficam vinculados a edital, anuncio e empresas participantes.
- Notificacoes sao lidas por usuario/empresa e deixam de aparecer como novas apos leitura.

## Indices recentes

O schema inclui indices para melhorar:

- feed da comunidade;
- perfil publico da empresa;
- comentarios ativos;
- curtidas e favoritos;
- noticias por status e vencimento;
- interesses por edital;
- vitrine de parceiros;
- matches por empresa;
- notificacoes por destinatario e entidade relacionada.
- fases e tarefas da montagem por posicao, responsavel, prazo e status.

## Criacao local

O backend atual aplica migracoes seguras ao iniciar. Para recriar manualmente o banco, use `schema.sql` com o `psql` do PostgreSQL.

Exemplo conceitual:

```powershell
psql -h localhost -p 5432 -U postgres -d licitahub_dev -f database/schema.sql
```

Use com cuidado em banco com dados reais.

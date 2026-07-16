# Testes automatizados

## Verificacao segura atual

Com o backend local ligado, execute:

```powershell
npm run test:smoke
```

O teste nao cria, altera ou remove registros. Ele valida disponibilidade do sistema e bloqueio de rotas sem sessao.

## Homologacao completa

Os fluxos que criam empresas, editais, consorcios e tarefas devem rodar somente em uma base de homologacao isolada. A massa deve conter um administrador da plataforma e tres empresas de teste, conforme o roteiro em `docs/ROTEIRO-DE-TESTES.md`.

Para criar a copia local sem alterar a base principal, abra `PREPARAR-HOMOLOGACAO.cmd`. Em seguida, abra `INICIAR-HOMOLOGACAO.cmd`; a versao de teste ficara em `http://127.0.0.1:8081`.

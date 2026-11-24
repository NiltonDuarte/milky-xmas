# Drops de atividade: números e variáveis

Referência de balanceamento dos dois drops, tirada do código.

Arquivos de origem: `src/internal/core/activity_drops.go` (constantes e regras de raridade), `core/gifts.go` (raridade do daily), `app/activity_drops.go` (contagem, cooldown, limites e tempo de vida), `adapter/discord/activity_drops.go` (mensagens, botões e fim do drop).

> Os valores com `Constante` têm nome no código. Os outros estão escritos direto na função indicada.

## 1. O que conta como atividade

- Só mensagens de **pessoas** (bots e webhooks não contam).
- Só em canais com os comandos ativados por `/setup channel`.
- A contagem é **por canal** e fica na memória (zera quando o bot reinicia).

## 2. Random Active Drop

| Variável | Valor | Onde |
| --- | --- | --- |
| Chance por mensagem | **1 em 20 (5%)** (`Constante RandomActiveDropRolls = 20`, aparece quando o sorteio dá 0) | `core.RandomActiveDropEligible` |
| Botões | **8** (2 linhas de 4), `Constante GiftDropSlots = 8` | `core` e Discord |
| Coletas por jogador | **1** por drop | `app.Claim` |
| Prêmio | **1 gift** | `app.Claim` |
| Vida do drop | **10 min** (`Constante GiftDropLifetime`) | `app/activity_drops.go` |
| Imagem | `grinchmax.png` | Discord |
| Botão aberto | fica **desabilitado** para todos, e cada botão só pode ser aberto uma vez | `app.Claim` e Discord |

**Raridade do gift**

| Jogador | Normal | Rare | Epic |
| --- | --- | --- | --- |
| Good deeds **abaixo de 50%** | 70% | 25% | 5% |
| Good deeds **de 50% ou mais** (`Constante GoodDeedsBoostThreshold = 50`) | **50%** | **35%** | **15%** |

- O sorteio é um número de 0 a 99. Sem bônus: 0-69 normal, 70-94 rare, 95-99 epic. Com bônus: 0-49 normal, 50-84 rare, 85-99 epic. (`core.ActiveDropRarity`)
- O gift não mexe no cooldown do daily nem na aura.

## 3. Gift Rain

| Variável | Valor | Onde |
| --- | --- | --- |
| Quando aparece | depois de **5 a 100 mensagens** no canal (sorteado de novo a cada chuva), `Constante GiftRainMinimumMessages = 5` e `GiftRainMessageRange = 96` (5 + 0 a 95) | `core.NextGiftRainThreshold` |
| Prioridade | na mensagem que dispara a chuva, o Random Active Drop **não** é sorteado | `app.Observe` |
| Botões | **8** (2 linhas de 4), todos equivalentes | Discord |
| Coletas até acabar | **10 no total**, de qualquer jogador (`Constante GiftRainClaimLimit = 10`) | `app.Claim` |
| Limite por jogador | **1 coleta por minuto** (`Constante GiftRainClaimCooldown`), vale para todas as chuvas juntas | `app.Claim` |
| Aviso de espera | "You cannot collect another gift yet.", **sem dizer o tempo** | Discord |
| Vida do drop | **10 min** (`Constante GiftDropLifetime`), se não chegar a 10 coletas antes | `app/activity_drops.go` |
| Imagem | `giftrain.webp` | Discord |
| Botões | **não** são desabilitados | Discord |

**Prêmio de cada coleta**

| Resultado | Chance |
| --- | --- |
| **Coal** (1 unidade) | **25%** (`Constante RainCoalPercent = 25`) |
| Gift normal | 75% × 70% = **52,5%** |
| Gift rare | 75% × 25% = **18,75%** |
| Gift epic | 75% × 5% = **3,75%** |

- Primeiro vem um sorteio de 0 a 99 para o Coal (0-24). Se não for Coal, um segundo sorteio usa as raridades do daily (70/25/5). O bônus de Good deeds **não** vale na chuva.
- O Coal não mexe na aura (a aura de Bad deeds só sobe ao **enviar** Coal com `/send gift`).

## 4. Fim do drop

- Quando o drop acaba, o bot **edita** a mensagem para `⌛ This drop has expired.`, tira os botões e lista cada coleta (jogador e prêmio, na ordem), ou diz "Nobody collected a gift."
- Os nomes aparecem como menção, **sem notificar**.
- O drop acaba por **tempo** (10 min) ou, na chuva, pela **10ª coleta**. A mensagem é editada uma vez só.
- Clique depois do fim: "This gift drop is no longer available."
- Os drops ficam só na memória. Se o bot reiniciar, a mensagem de um drop aberto **não** é editada.

## 5. Respostas privadas ao clicar

| Situação | Resposta |
| --- | --- |
| Gift | "You collected a rare gift." |
| Coal (só na chuva) | "You collected 🌑 Coal." |
| Já coletou neste drop (ativo) | "You already collected a gift from this drop." |
| Botão já aberto por outro (ativo) | "Someone already opened that gift." |
| Menos de 1 min desde a última coleta (chuva) | "You cannot collect another gift yet." |
| Drop acabou ou expirou | "This gift drop is no longer available." |

## 6. Limites técnicos

- Um jogador nunca passa de `MaxInt64` gifts ou Coal (daria erro, não acontece na prática).
- Um drop esquecido (sem timer) é apagado da memória **1 min depois** de passar dos 10 min.
- Cada clique salva o prêmio no banco antes de contar. Dois cliques ao mesmo tempo do mesmo jogador dão **um** prêmio só.

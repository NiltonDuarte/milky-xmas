# Drops de atividade: números e variáveis

Referência de balanceamento dos dois drops. Os valores abaixo são os **padrões**. Para ajustá-los, edite [`tuning.yaml`](../tuning.yaml) e inicie o bot com `MILKY_TUNING_CONFIG=tuning.yaml` (veja a seção 7).

Arquivos de origem: `src/internal/core/tuning.go` (padrões e validação), `core/activity_drops.go` e `core/gifts.go` (regras de raridade), `app/activity_drops.go` (contagem, cooldown, limites e tempo de vida), `adapter/discord/activity_drops.go` (mensagens, botões e fim do drop).

> Os valores `Constante` são os padrões com nome em `core/tuning.go`. A coluna "Chave no YAML" diz onde cada número é ajustado.

## 1. O que conta como atividade

- Só mensagens de **pessoas** (bots e webhooks não contam).
- Só em canais com os comandos ativados por `/setup channel`.
- A contagem é **por canal** e fica na memória (zera quando o bot reinicia).

## 2. Random Active Drop

| Variável | Valor | Onde |
| --- | --- | --- |
| Chance por mensagem | **1 em 20 (5%)** (`drops.random_active.one_in`, aparece quando o sorteio dá 0) | `core.DropTuning.RandomActiveDropEligible` |
| Botões | **8** (2 linhas de 4), fixo (não ajustável) | `core` e Discord |
| Coletas por jogador | **1** por drop | `app.Claim` |
| Prêmio | **1 gift** | `app.Claim` |
| Vida do drop | **10 min** (`drops.lifetime`) | `app/activity_drops.go` |
| Imagem | `grinchmax.png` | Discord |
| Botão aberto | fica **desabilitado** para todos, e cada botão só pode ser aberto uma vez | `app.Claim` e Discord |

**Raridade do gift**

| Jogador (`gifts.daily_rarity` e `gifts.good_deeds_boost.rarity`) | Normal | Rare | Epic |
| --- | --- | --- | --- |
| Fora da Nice list (aura **abaixo de +25**) | 70% | 25% | 5% |
| Nice list, aura **+25 ou mais** (`aura.nice_at`) | **50%** | **35%** | **15%** |

- O sorteio é um número de 0 a 99. Sem bônus: 0-69 normal, 70-94 rare, 95-99 epic. Com bônus: 0-49 normal, 50-84 rare, 85-99 epic. (`core.GiftTuning.ActiveDropRarity`)
- O gift não mexe no cooldown do daily nem na aura.

## 3. Gift Rain

| Variável | Valor | Onde |
| --- | --- | --- |
| Quando aparece | depois de **5 a 100 mensagens** no canal (sorteado de novo a cada chuva), `drops.gift_rain.min_messages` e `max_messages` | `core.DropTuning.NextGiftRainThreshold` |
| Prioridade | na mensagem que dispara a chuva, o Random Active Drop **não** é sorteado | `app.Observe` |
| Botões | **8** (2 linhas de 4), todos equivalentes | Discord |
| Coletas até acabar | **10 no total**, de qualquer jogador (`drops.gift_rain.claims_to_end`) | `app.Claim` |
| Limite por jogador | **1 coleta por minuto** (`drops.gift_rain.claim_cooldown`), vale para todas as chuvas juntas | `app.Claim` |
| Aviso de espera | "You cannot collect another gift yet.", **sem dizer o tempo** | Discord |
| Vida do drop | **10 min** (`drops.lifetime`), se não chegar a 10 coletas antes | `app/activity_drops.go` |
| Imagem | `giftrain.webp` | Discord |
| Botões | **não** são desabilitados | Discord |

**Prêmio de cada coleta**

| Resultado | Chance |
| --- | --- |
| **Coal** (1 unidade) | **25%** (`drops.gift_rain.coal_percent`) |
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

## 7. Onde ajustar cada número

Arquivo: [`tuning.yaml`](../tuning.yaml), apontado por `MILKY_TUNING_CONFIG` (Discord e chat local). O arquivo é lido uma vez, na inicialização, então reinicie o bot para aplicar. Chaves omitidas ficam no padrão, e chaves desconhecidas ou valores fora da faixa impedem o bot de iniciar, com todos os problemas listados.

| Número | Chave no YAML | Padrão | Regra |
| --- | --- | --- | --- |
| Vida do drop | `drops.lifetime` | `10m` | maior que zero |
| Chance do Random Active Drop | `drops.random_active.one_in` | `20` (1 em 20) | 1 ou mais |
| Mínimo de mensagens da chuva | `drops.gift_rain.min_messages` | `5` | 1 ou mais |
| Máximo de mensagens da chuva | `drops.gift_rain.max_messages` | `100` | igual ou maior que o mínimo |
| Coletas que encerram a chuva | `drops.gift_rain.claims_to_end` | `10` | 1 ou mais |
| Espera entre coletas da chuva | `drops.gift_rain.claim_cooldown` | `1m` | 0 ou mais |
| Chance de Coal na chuva | `drops.gift_rain.coal_percent` | `25` | 0 a 100 |
| Raridade do daily e da chuva | `gifts.daily_rarity.{normal,rare,epic}` | `70`, `25`, `5` | cada um de 0 a 100, soma 100 |
| Aura da Nice list | `aura.nice_at` | `25` | 1 a 50 |
| Raridade com bônus | `gifts.good_deeds_boost.rarity.{normal,rare,epic}` | `50`, `35`, `15` | soma 100 |

- `gifts.daily_rarity` vale também para `/daily claim`.
- A chuva sorteia o limite entre `min_messages` e `max_messages`, com os dois incluídos. Os `5` e `96` da seção 3 viram `5` e `100`.
- Não são ajustáveis: os 8 botões por drop e o tamanho das cartelas.

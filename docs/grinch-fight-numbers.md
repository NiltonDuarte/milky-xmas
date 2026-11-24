# Luta do Grinch: números e variáveis

Referência de balanceamento, tirada do código. Os valores internos usam **centésimos de HP** (10000 = 100% de HP). Aqui tudo está em % de HP, salvo onde indicado.

Arquivos de origem: `src/internal/core/grinch.go` (regras), `core/grinch_rewards.go` (recompensas), `core/grinch_scores.go` (placar), `app/grinch.go` (sorteios e orquestração), `adapter/discord/grinch.go` (botões e mensagem).

> Quase todos os números abaixo estão escritos direto no código, sem constante com nome. Só os marcados com `Constante` têm nome. Para mudar um valor, procure-o em `core/grinch.go`.

## 1. Aparecimento do boss

| Variável | Valor | Onde |
| --- | --- | --- |
| HP total do boss | **100%** (10000 centésimos) | `app/grinch.go`, `Spawn` |
| Chance de aparecer sozinho | **10%** por rolagem (`roll < 10` de 100) | `core/grinch.go`, `CanSpawn` |
| Frequência da rolagem | **1 por minuto** | `adapter/discord/grinch.go`, `run` |
| Janela de atividade exigida | alguém falou no canal do boss nos **últimos 10 min** | `CanSpawn` |
| Intervalo mínimo depois de uma luta | **30 min** após o fim da anterior | `CanSpawn` |
| Só uma luta por vez | sim, enquanto houver uma ativa | `CanSpawn` |
| Invocação por moderador (`/grinch summon`) | ignora a chance, a atividade e os 30 min, mas não cria uma segunda luta | `CanSpawn` |

## 2. Inscrição (lobby)

| Variável | Valor | Onde |
| --- | --- | --- |
| Tempo para se inscrever | **1 min** (`Constante GrinchJoinDuration`) | `core/grinch.go` |
| Início antecipado | com **3 inscritos** a luta começa na hora | `Move`, fase lobby |
| Ninguém se inscreve | a luta é cancelada, sem recompensa | `Advance` |
| Entrar depois | permitido até o fim do relógio de combate (a luta já começou) | `CanJoin` |

## 3. Fases e dano

### Fase 1: Snowballs (☃️ Attack)
- Cada clique de um inscrito tira **1% de HP** do boss.
- Termina quando o boss chega a **75%**, ou seja, depois de **25 cliques** no total, de qualquer jogador.
- **Não há relógio** nesta fase.
- Quando chega a 75%, o relógio de **1 min** começa (`TimedAt`) e a fase 2 abre. O número de jogadores nesse instante fica registrado (`DefensePlayers`).

### Relógio de combate
- **1 min** (`Constante GrinchDuration`), contado a partir do boss chegar a 75%.
- Cobre as fases 2, 3 e 4. Se acabar antes de o boss morrer, o Grinch **escapa**.

### Fase 2: Defesa (🎄 / ❄️)
- Tabuleiro de 2 linhas de 4 botões (8 casas). Cada casa é 🎄 ou ❄️, sorteadas de novo a cada clique, e sempre há pelo menos uma de cada.
- **🎄 (defesa certa):** conta como uma defesa bem-sucedida. Cada defesa tira `30 ÷ nº de jogadores` % de HP do boss.
- Ao todo são necessárias **tantas defesas quanto jogadores** inscritos. Elas tiram **30%** de HP no total, e o boss vai de **75% para 45%**.
- **❄️ (errou):** o jogador perde **10% do HP máximo dele**. Com 10 erros ele fica com 0 e não pode mais agir.
- Se **todos** os jogadores chegarem a 0 de HP, o resultado é **derrota** na hora.
- O tabuleiro é trocado a cada clique, e cliques de um tabuleiro velho são recusados.

### Fase 3: Ataque (🧊 / ☃️)
- Também 2 linhas de 4 botões, com pelo menos um 🧊 e um ☃️. O tabuleiro é sorteado **uma vez**, quando a fase começa, e **não muda** durante a fase (diferente da defesa e do ataque final).
- **☃️:** tira **5%** de HP do boss.
- **🧊:** tira **5,25%** de HP do boss.
- A fase termina quando o boss chega a **5%**. Ele sai de 45%, então faltam **40%**, o que dá no mínimo 8 cliques (os últimos cliques são cortados para não passar de 5%).
- Os jogadores não perdem HP nesta fase.

### Fase 4: Ataque final (⚔️ / 🎁 / ❄️)
- Uma linha de 5 botões: 4 casas que são ⚔️ ou 🎁 (sorteadas a cada clique, sempre com pelo menos um de cada) e um botão fixo ❄️ no fim.
- **⚔️:** tira **1%** de HP do boss e dá **2 pontos de dano** ao jogador.
- **🎁:** não tira HP. Dá **1 a 5 pontos de presente** (sorteio uniforme) ao jogador.
- **❄️:** não tira HP. Dá **1 Snowball** no inventário do jogador na hora.
- O boss está com 5%, então **5 cliques em ⚔️** (de qualquer jogador) o derrotam. Isso é a **vitória**.

## 4. Fim da luta e recompensas

| Resultado | Quando | Recompensa |
| --- | --- | --- |
| **Vitória** | boss chega a 0% | só o que já foi ganho na hora: Snowballs do ❄️ final. Os pontos de presente entram no placar. |
| **Fuga** | o relógio de 1 min acaba | cada jogador inscrito recebe **1 Coal** e **Yellow Snowballs = nº de jogadores inscritos** |
| **Derrota** | todos os jogadores chegam a 0 de HP | igual à fuga |
| **Cancelada** | ninguém se inscreveu | nada |

- Na fuga e na derrota, quem tinha se inscrito recebe a recompensa mesmo que tenha ficado com 0 de HP.
- A imagem do Max aparece na mensagem final quando o Grinch **sobrevive** (fuga ou derrota).

## 5. Placares

**Dano** (em pontos, mostrado com 2 casas):

| Ação | Pontos |
| --- | --- |
| Clique de Snowball (fase 1) | 1 |
| Defesa certa (fase 2) | `30 ÷ nº de jogadores` |
| Ataque ☃️ (fase 3) | 5 |
| Ataque 🧊 (fase 3) | 5,25 (pode ser menos no último golpe) |
| ⚔️ (fase 4) | **2** (embora tire só 1% de HP) |

- **Presentes:** soma dos pontos 🎁 (1 a 5 cada).
- Os dois placares só listam jogadores inscritos. Desempate pelo ID do jogador.

## 6. Resumo de uma luta

Para N jogadores inscritos:

1. 25 cliques de ☃️ levam o boss de 100% a 75%. O relógio de 1 min começa.
2. N defesas certas levam o boss a 45% (cada erro custa 10% do HP máximo do jogador).
3. Cerca de 8 ataques (5% ou 5,25% cada) levam o boss a 5%.
4. 5 cliques em ⚔️ derrotam o boss. Cada 🎁 e cada ❄️ clicados gastam um clique sem dano.

Total mínimo: **25 + N + 8 + 5 cliques**, tudo dentro de 1 min depois dos primeiros 25.

## 7. Números que não pertencem à luta mas aparecem nela

- Cada clique usa um ID único, então um clique repetido pela mesma interação não conta duas vezes.
- Um jogador sem HP (10 erros na defesa) não pode mais clicar em nenhuma fase.

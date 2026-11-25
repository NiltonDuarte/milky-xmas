# Luta do Grinch: números e variáveis

Referência de balanceamento. Os valores abaixo são os **padrões**. Para ajustá-los, edite [`tuning.yaml`](../tuning.yaml) e inicie o bot com `MILKY_TUNING_CONFIG=tuning.yaml` (veja a seção 8). Os valores internos usam **centésimos de HP** (10000 = 100% de HP). Aqui e no YAML tudo está em % de HP, com até 2 casas decimais.

Arquivos de origem: `src/internal/core/tuning.go` (padrões e validação), `core/grinch.go` (regras), `core/grinch_rewards.go` (recompensas), `core/grinch_scores.go` (placar), `app/grinch.go` (sorteios e orquestração), `adapter/discord/grinch.go` (botões e mensagem).

> Os textos da luta que citam um número (por exemplo "75 HP", "0.84 damage" e "n/3 to start early") seguem o valor configurado.

## 1. Aparecimento do boss

| Variável | Valor | Onde |
| --- | --- | --- |
| HP total do boss | **100%** (10000 centésimos), fixo | `core.GrinchBossHP` |
| Chance de aparecer sozinho | **10%** por rolagem (`grinch.spawn.chance_percent`) | `core/grinch.go`, `CanSpawn` |
| Frequência da rolagem | **1 por minuto** (`grinch.spawn.check_interval`) | `adapter/discord/grinch.go`, `run` |
| Janela de atividade exigida | alguém falou no canal do boss nos **últimos 10 min** (`grinch.spawn.activity_window`) | `CanSpawn` |
| Intervalo mínimo depois de uma luta | **30 min** após o fim da anterior (`grinch.spawn.cooldown_after_fight`) | `CanSpawn` |
| Só uma luta por vez | sim, enquanto houver uma ativa | `CanSpawn` |
| Invocação por moderador (`/grinch summon`) | ignora a chance, a atividade e os 30 min, mas não cria uma segunda luta | `CanSpawn` |

## 2. Start (lobby)

O anúncio tem um botão **Start**. O primeiro clique começa a contagem regressiva, e quem aperta Start entra na luta.

| Variável | Valor | Onde |
| --- | --- | --- |
| Tempo para o primeiro Start | **1 min** depois do anúncio (`grinch.registration.start_window`). Sem clique, a luta é cancelada, sem recompensa | `LobbyDeadline`, `Advance` |
| Contagem regressiva | **1 min** a partir do primeiro Start (`grinch.registration.duration`) | `LobbyDeadline`, `Move` |
| Início antecipado | com **3 jogadores** que apertaram Start a luta começa na hora (`grinch.registration.early_start_players`) | `Move`, fase lobby |
| Participar | não existe "entrar na luta": com a luta em andamento qualquer pessoa pode clicar nos botões de ataque e passa a participar no primeiro clique aceito. Um clique recusado não inscreve ninguém | `Move` |

## 3. Fases e dano

### Fase 1: Snowballs (☃️ Attack)
- Cada clique tira **4,17% de HP** do boss (`grinch.snowballs.damage`), qualquer que seja o número de jogadores.
- Termina quando o boss chega a **75%** (`grinch.snowballs.boss_hp_after`), ou seja, depois de **6 cliques** no total, de qualquer jogador (o último é cortado no marco).
- O grupo tem **2 min** para chegar a 75%, contados do início do combate (`grinch.snowballs.duration`). Se o tempo acabar antes, o **boss vence** (a luta termina como fuga, com as recompensas de derrota).
- Quando chega a 75%, o relógio de **1 min** começa (`TimedAt`) e a fase 2 abre. O número de jogadores nesse instante, no máximo 6, fica registrado (`DefensePlayers`).

### Relógio de combate
- **1 min** (`grinch.fight_duration`), contado a partir do boss chegar a 75%.
- Cobre as fases 2, 3 e 4. Se acabar antes de o boss morrer, o Grinch **escapa**.

### Fase 2: Defesa (🎄 / ❄️)
- Tabuleiro de 2 linhas de 4 botões (8 casas). Cada casa é 🎄 ou ❄️, sorteadas de novo a cada clique, e sempre há pelo menos uma de cada.
- **🎄 (defesa certa):** conta como uma defesa bem-sucedida. Cada defesa tira `30 ÷ defesas necessárias` % de HP do boss.
- Ao todo é necessária **uma defesa por jogador, no máximo 6** (`grinch.defense.max_defenses`). Elas tiram **30%** de HP no total, e o boss vai de **75% para 45%**.
- **❄️ (errou):** o jogador perde **10% do HP máximo dele** (`grinch.defense.miss_hp_loss_percent`). Com 10 erros ele fica com 0 e não pode mais agir.
- Se **todos** os jogadores chegarem a 0 de HP, o resultado é **derrota** na hora.
- O tabuleiro é trocado a cada clique, e cliques de um tabuleiro velho são recusados.

### Fase 3: Ataque (🧊 / ☃️)
- Também 2 linhas de 4 botões, com pelo menos um 🧊 e um ☃️. O tabuleiro é sorteado **uma vez**, quando a fase começa, e **não muda** durante a fase (diferente da defesa e do ataque final).
- **☃️:** tira **6,67%** de HP do boss (`grinch.attack.damage`).
- **🧊:** tira **7%** de HP do boss (`grinch.attack.heavy_damage`).
- A fase termina quando o boss chega a **5%** (`grinch.attack.boss_hp_after`). Ele sai de 45%, então faltam **40%**, o que dá 6 cliques (o último é cortado para não passar de 5%).
- Os jogadores não perdem HP nesta fase.

### Fase 4: Ataque final (⚔️ / 🎁 / ❄️)
- Uma linha de 5 botões: 4 casas que são ⚔️ ou 🎁 (sorteadas a cada clique, sempre com pelo menos um de cada) e um botão fixo ❄️ no fim.
- **⚔️:** tira **0,84%** de HP do boss (`grinch.final_attack.hit_damage`) e dá **2 pontos** ao jogador (`hit_score`). Os pontos ficam guardados por jogador para uso futuro e **não aparecem no ranking de dano**, que mostra só o dano causado. **Só o ataque final dá pontos**: snowball, defesa e ataque não dão.
- **🎁:** não tira HP. Dá **1 a 5 pontos de presente** (sorteio uniforme, `gift_points.min` e `max`) ao jogador.
- **❄️:** não tira HP. Dá **1 Snowball** no inventário do jogador na hora (`snowballs_reward`).
- O boss está com 5%, então **6 cliques em ⚔️** (de qualquer jogador) o derrotam. Isso é a **vitória**.

## 4. Fim da luta e recompensas

| Resultado | Quando | Recompensa |
| --- | --- | --- |
| **Vitória** | boss chega a 0% | só o que já foi ganho na hora: Snowballs do ❄️ final. Os pontos de presente entram no placar. |
| **Fuga** | o relógio de 1 min acaba | cada jogador inscrito recebe **1 Coal** (`loss_rewards.coal`) e **Yellow Snowballs = nº de jogadores inscritos** (`loss_rewards.yellow_snowballs_per_group_member` por inscrito) |
| **Derrota** | todos os jogadores chegam a 0 de HP | igual à fuga |
| **Cancelada** | ninguém se inscreveu | nada |

- Na fuga e na derrota, quem tinha se inscrito recebe a recompensa mesmo que tenha ficado com 0 de HP.
- A imagem do Max aparece na mensagem final quando o Grinch **sobrevive** (fuga ou derrota).

## 5. Placares

**Dano** (em pontos, mostrado com 2 casas):

| Ação | Pontos |
| --- | --- |
| Clique de Snowball (fase 1) | 4,17 |
| Defesa certa (fase 2) | `30 ÷ defesas necessárias` |
| Ataque ☃️ (fase 3) | 6,67 |
| Ataque 🧊 (fase 3) | 7 |
| ⚔️ (fase 4) | 0,84 |

O último golpe de cada fase pode valer menos, porque é cortado no marco.

- **Presentes:** soma dos pontos 🎁 (1 a 5 cada).
- Os dois placares só listam jogadores inscritos. Desempate pelo ID do jogador.

## 6. Resumo de uma luta

Para N jogadores:

1. 6 cliques de ☃️ levam o boss de 100% a 75%. O relógio de 1 min começa.
2. N defesas certas (no máximo 6) levam o boss a 45% (cada erro custa 10% do HP máximo do jogador).
3. 6 ataques (6,67% ou 7% cada) levam o boss a 5%.
4. 6 cliques em ⚔️ derrotam o boss. Cada 🎁 e cada ❄️ clicados gastam um clique sem dano.

Total mínimo: **6 + min(N, 6) + 6 + 6 cliques**, com o relógio de 1 min depois dos primeiros 6.

## 7. Números que não pertencem à luta mas aparecem nela

- Cada clique usa um ID único, então um clique repetido pela mesma interação não conta duas vezes.
- Um jogador sem HP (10 erros na defesa) não pode mais clicar em nenhuma fase.

## 8. Onde ajustar cada número

Arquivo: [`tuning.yaml`](../tuning.yaml), apontado por `MILKY_TUNING_CONFIG` (Discord e chat local). O arquivo é lido uma vez, na inicialização, então reinicie o bot para aplicar. Reinicie com o bot sem luta aberta: uma luta já salva no banco continuaria com os números novos.

| Número | Chave no YAML | Padrão |
| --- | --- | --- |
| Intervalo entre rolagens de aparição | `grinch.spawn.check_interval` | `1m` |
| Chance de aparecer (%) | `grinch.spawn.chance_percent` | `10` |
| Janela de atividade no canal | `grinch.spawn.activity_window` | `10m` |
| Intervalo mínimo após uma luta | `grinch.spawn.cooldown_after_fight` | `30m` |
| Tempo de inscrição | `grinch.registration.duration` | `1m` |
| Inscritos que iniciam a luta na hora | `grinch.registration.early_start_players` | `3` |
| Relógio de combate | `grinch.fight_duration` | `1m` |
| Dano por clique de snowball (% de HP) | `grinch.snowballs.damage` | `4.17` |
| HP do boss ao fim da fase 1 | `grinch.snowballs.boss_hp_after` | `75` |
| HP do boss ao fim da defesa | `grinch.defense.boss_hp_after` | `45` |
| Máximo de defesas necessárias | `grinch.defense.max_defenses` | `6` |
| HP perdido por erro na defesa (% do máximo do jogador) | `grinch.defense.miss_hp_loss_percent` | `10` |
| Dano do ☃️ | `grinch.attack.damage` | `6.67` |
| Dano do 🧊 | `grinch.attack.heavy_damage` | `7` |
| HP do boss ao fim do ataque | `grinch.attack.boss_hp_after` | `5` |
| Dano do ⚔️ final | `grinch.final_attack.hit_damage` | `0.84` |
| Pontos do ⚔️ final (guardados) | `grinch.final_attack.hit_score` | `2` |
| Pontos de presente do 🎁 | `grinch.final_attack.gift_points.min` e `max` | `1` e `5` |
| Snowballs do ❄️ final | `grinch.final_attack.snowballs_reward` | `1` |
| Coal por inscrito na fuga ou derrota | `grinch.loss_rewards.coal` | `1` |
| Yellow Snowballs por membro do grupo | `grinch.loss_rewards.yellow_snowballs_per_group_member` | `1` |

Regras de validação (o bot não inicia se alguma falhar):

- Os marcos de HP precisam cair de 100 até 0: `0 < attack.boss_hp_after < defense.boss_hp_after < snowballs.boss_hp_after < 100`.
- O dano total da defesa é `snowballs.boss_hp_after - defense.boss_hp_after`, dividido entre as defesas necessárias (hoje `75 - 45 = 30`).
- Danos e HP aceitam até 2 casas decimais e precisam ser maiores que zero. O último golpe de cada fase é cortado no marco, então o dano não precisa dividir a distância em partes iguais.
- Durações maiores que zero (a janela de atividade e o intervalo após a luta podem ser `0s`). `early_start_players` é 1 ou mais. `chance_percent` vai de 0 a 100. `miss_hp_loss_percent` vai de 1 a 100. `gift_points`: `1 <= min <= max`. Recompensas são 0 ou mais.
- Não são ajustáveis: o HP total (100%), os 8 botões dos tabuleiros e a linha final de 5 botões.

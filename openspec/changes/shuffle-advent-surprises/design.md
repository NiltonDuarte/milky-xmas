## Decisions

- **Derived orders.** Both shuffles come from a hash of the server, player, year, and purpose, so they stay the same without being stored. The test window uses its own year key, so December deals new orders.
- **Missed doors stay open.** The product owner chose that a door can be opened on its day or any later day through December 24, so several missed doors can be opened at once.
- **Opened doors outlive the window.** An opened door shows its surprise even after December 24.
- **Embed for the surprise.** A recipe can pass the 2000 characters of a plain message, so it is sent as an embed, up to 4096.
- **Local chat.** It has no buttons, so `/advent open <day>` stands in for a click there only.
- **Content order is not secret.** Anyone who knows the code and a player's IDs could compute it, which does not matter for recipes.

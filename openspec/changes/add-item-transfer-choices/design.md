# Design

## Decisions

- Use an optional Discord autocomplete field to suggest Normal and, when owned, Yellow. The shared service validates the final selection and checks stock again at commit time.
- Reuse the paired compare-and-save inventory operation for Yellow Snowball and Coal transfers. Their domain rules preserve unrelated balances and aura.
- Preserve the existing normal throw and gift behavior. The local chat accepts an optional snowball type and an item choice for sends.

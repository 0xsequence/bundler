# Entrypoint and mock environment

## Deployment

Create `.env` with the following:

```
PRIVATE_KEY=0x...
```

Run the deploy command via make:

```
make deploy
```

Contracts will be deployed to:

| Contract             | Address                                      |
| -------------------- | -------------------------------------------- |
| `OperationValidator` | `0x6CDDB903C49CF49e52b853C7a83F6A79E17a0BaA` |
| `TemporalRegistry`   | `0x292D731B485c6B6fa561Afb663D56A397c319Bfd` |

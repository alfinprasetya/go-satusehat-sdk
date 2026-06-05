# SATUSEHAT FHIR Postman Guidelines

Use this directory to store Postman assets copied or adapted from the official SATUSEHAT FHIR documentation. These files are intended as implementation guidelines for SDK development, request mapping, and manual API exploration.

## Directory layout

```text
docs/postman
├── collections
│   └── README.md
├── environments
│   └── README.md
└── README.md
```

## Recommended workflow

1. Export or download the SATUSEHAT FHIR R4 Postman collection from the official documentation.
2. Save collection files in `collections/` using descriptive names, for example `satusehat-fhir-r4.postman_collection.json`.
3. Save non-secret environment templates in `environments/`, for example `satusehat-staging.postman_environment.json`.
4. Keep real credentials out of the repository. Use Postman's current values or local secret storage for client IDs, client secrets, organization IDs, and tokens.
5. When adding SDK support for a FHIR resource, reference the matching Postman request as the source guideline for endpoint paths, query parameters, request bodies, and expected responses.

## Notes

- Prefer official SATUSEHAT examples over hand-written examples when both exist.
- Keep collection exports readable where possible by avoiding unnecessary Postman metadata churn.
- Do not commit generated access tokens, refresh tokens, client secrets, or production identifiers.

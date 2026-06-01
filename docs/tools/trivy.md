# trivy

**What it checks:** Security scanner — finds vulnerabilities in dependencies, containers, and IaC.

## Without prettyout

### Default output

```
Report Summary

┌──────────────────┬──────┬─────────────────┬─────────┐
│      Target      │ Type │ Vulnerabilities │ Secrets │
├──────────────────┼──────┼─────────────────┼─────────┤
│ requirements.txt │ pip  │       16        │    -    │
└──────────────────┴──────┴─────────────────┴─────────┘
Legend:
- '-': Not scanned
- '0': Clean (no security findings detected)


requirements.txt (pip)
======================
Total: 16 (UNKNOWN: 0, LOW: 1, MEDIUM: 6, HIGH: 6, CRITICAL: 3)

┌─────────┬────────────────┬──────────┬────────┬───────────────────┬────────────────────────┬──────────────────┐
│ Library │ Vulnerability  │ Severity │ Status │ Installed Version │     Fixed Version      │      Title       │
├─────────┼────────────────┼──────────┼────────┼───────────────────┼────────────────────────┼──────────────────┤
│ django  │ CVE-2019-19844 │ CRITICAL │ fixed  │ 2.0               │ 1.11.27, 2.2.9, 3.0.1  │ Django: crafted  │
│         │                │          │        │                   │                        │ email address    │
│         │                │          │        │                   │                        │ allows account   │
│         │                │          │        │                   │                        │ takeover         │
│         ├────────────────┤          │        │                   ├────────────────────────┼──────────────────┤
│         │ CVE-2020-7471  │          │        │                   │ 1.11.28, 2.2.10, 3.0.3 │ django: potential│
│         │                │          │        │                   │                        │ SQL injection    │
...
```

### JSON (what CI/CD sees)

```json
{
  "SchemaVersion": 2,
  "ArtifactName": ".",
  "ArtifactType": "filesystem",
  "Results": [
    {
      "Target": "requirements.txt",
      "Class": "lang-pkgs",
      "Type": "pip",
      "Packages": [
        {
          "Name": "django",
          "Version": "2.0",
          ...
        }
      ],
      "Vulnerabilities": [
        {
          "VulnerabilityID": "CVE-2019-19844",
          "PkgName": "django",
          ...
        },
        ...
      ]
    }
  ]
}
```

## With prettyout

### Group by rule (default — grouped by CVE)

```
CRITICAL (3 vulnerabilities)
  CVE-2019-19844 — django 2.0 → fix: 1.11.27, 2.2.9, 3.0.1
  CVE-2020-7471 — django 2.0 → fix: 1.11.28, 2.2.10, 3.0.3
  CVE-2025-64459 — django 2.0 → fix: 5.2.8, 5.1.14, 4.2.26
────────────────────────────────────────────────
HIGH (6 vulnerabilities)
  CVE-2018-6188 — django 2.0 → fix: 2.0.2, 1.11.10
  CVE-2019-3498 — django 2.0 → fix: 1.11.18, 2.0.10, 2.1.5
  CVE-2019-6975 — django 2.0 → fix: 1.11.19, 2.0.11, 2.1.6
  CVE-2022-36359 — django 2.0 → fix: 3.2.15, 4.0.7
  CVE-2025-57833 — django 2.0 → fix: 4.2.24, 5.1.12, 5.2.6
  CVE-2025-64458 — django 2.0 → fix: 5.2.8, 5.1.14, 4.2.26
────────────────────────────────────────────────
MEDIUM (6 vulnerabilities)
  CVE-2018-14574 — django 2.0 → fix: 2.0.8, 1.11.15
  CVE-2018-7536 — django 2.0 → fix: 2.0.3, 1.11.11, 1.8.19
  CVE-2019-11358 — django 2.0 → fix: 2.1.9, 2.2.2
  CVE-2021-33203 — django 2.0 → fix: 2.2.24, 3.1.12, 3.2.4
  CVE-2024-45231 — django 2.0 → fix: 5.1.1, 5.0.9, 4.2.16
  CVE-2025-48432 — django 2.0 → fix: 5.2.2, 5.1.10, 4.2.22
────────────────────────────────────────────────
LOW (1 vulnerability)
  CVE-2018-7537 — django 2.0 → fix: 2.0.3, 1.11.11, 1.8.19
────────────────────────────────────────────────
16 vulnerabilities · 4 severity levels
```

### Group by file (grouped by package file)

```
requirements.txt
  CRITICAL CVE-2019-19844 — django 2.0 → fix: 1.11.27, 2.2.9, 3.0.1
  CRITICAL CVE-2020-7471 — django 2.0 → fix: 1.11.28, 2.2.10, 3.0.3
  CRITICAL CVE-2025-64459 — django 2.0 → fix: 5.2.8, 5.1.14, 4.2.26
  HIGH     CVE-2018-6188 — django 2.0 → fix: 2.0.2, 1.11.10
  HIGH     CVE-2019-3498 — django 2.0 → fix: 1.11.18, 2.0.10, 2.1.5
  HIGH     CVE-2019-6975 — django 2.0 → fix: 1.11.19, 2.0.11, 2.1.6
  HIGH     CVE-2022-36359 — django 2.0 → fix: 3.2.15, 4.0.7
  HIGH     CVE-2025-57833 — django 2.0 → fix: 4.2.24, 5.1.12, 5.2.6
  HIGH     CVE-2025-64458 — django 2.0 → fix: 5.2.8, 5.1.14, 4.2.26
  MEDIUM   CVE-2018-14574 — django 2.0 → fix: 2.0.8, 1.11.15
  MEDIUM   CVE-2018-7536 — django 2.0 → fix: 2.0.3, 1.11.11, 1.8.19
  MEDIUM   CVE-2019-11358 — django 2.0 → fix: 2.1.9, 2.2.2
  MEDIUM   CVE-2021-33203 — django 2.0 → fix: 2.2.24, 3.1.12, 3.2.4
  MEDIUM   CVE-2024-45231 — django 2.0 → fix: 5.1.1, 5.0.9, 4.2.16
  MEDIUM   CVE-2025-48432 — django 2.0 → fix: 5.2.2, 5.1.10, 4.2.22
  LOW      CVE-2018-7537 — django 2.0 → fix: 2.0.3, 1.11.11, 1.8.19
────────────────────────────────────────────────
16 vulnerabilities · 1 target
```

## What prettyout improves

- **Consistent format**: trivy table output is unique to trivy → prettyout produces the same structure as all other supported tools
- **Severity ordering**: raw output mixes severities → prettyout groups by severity in CRITICAL → HIGH → MEDIUM → LOW → UNKNOWN order
- **Fix availability surfaced**: raw output buries "no fix available" in the table → prettyout shows it explicitly in the vulnerability entry
- **Rule grouping**: raw output groups by package file/image → group-by-rule (CVE) shows all packages affected by the same vulnerability
- **Summary line**: raw output ends with a table footer → `N issues · M rules · K files` is clean and consistent

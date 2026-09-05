# Spidey Versioning Strategy

Spidey uses a modified Calendar Versioning (CalVer) approach that remains compatible with Go's Semantic Versioning (SemVer) requirements.

## Format: `v[Stage].[Year].[MonthDate]`

Example: `v0.26.711`

### Breakdown:
1. **Stage (`0` or `1+`)**: 
   - `0` means the framework is in **Beta** / Pre-release.
   - `1`, `2`, etc., means it is a stable, production-ready release.
2. **Year (`26`)**: 
   - The 2-digit year (e.g., `26` for 2026).
3. **MonthDate (`0711`)**: 
   - The month and date combined. `7` for July, `11` for the 11th. (e.g., November 5th would be `1105`).

### Examples
* `v0.26.0711` -> Beta release on July 11th, 2026
* `v0.26.1205` -> Beta release on December 5th, 2026
* `v1.27.0115` -> First Stable release on January 15th, 2027

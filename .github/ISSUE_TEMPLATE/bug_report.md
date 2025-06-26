---
name: Bug report
about: Create a report to help us improve
title: ''
labels: ''
assignees: ''
---

### Describe the bug
A clear and concise description of what the bug is.

### To Reproduce
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

### Expected behavior
A clear and concise description of what you expected to happen.

### Screenshots
If applicable, add screenshots to help explain your problem.

### Your Kiosk version
This can be found on the startup banner in the terminal, in the browser (via the `<meta name="version" content={ KioskVersion } />` tag or in the browser console), or by visiting `/about`.

### Your Kiosk installation
- Docker
- Binary

### Your Kiosk configuration (ENV / config file)
Please remove any sensitive data (e.g. Immich API key or URL) before sharing your configuration.

**To generate a sanitised copy automatically:**

1. Enable one of the following debug flags
    - `KIOSK_DEBUG: true`
    - `KIOSK_DEBUG_VERBOSE: true`
    - `debug: true` **or** `debug_verbose: true` in `config.yaml`
2. Navigate to `/config` in your browser – a redacted YAML version will be displayed.

### Any params passed to the URL used to access Kiosk
- http://****/?show_time=true&time_format=12

### Desktop
- OS: [e.g. iOS]
- Browser [e.g. chrome, safari]
- Version [e.g. 22]

### Smartphone
- Device: [e.g. iPhone6]
- OS: [e.g. iOS8.1]
- Browser [e.g. stock browser, safari]
- Version [e.g. 22]

### Additional context
Add any other context about the problem here.

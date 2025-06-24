---
name: Bug report
about: Create a report to help us improve
title: ''
labels: ''
assignees: ''
---

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Your Kiosk version**
This can be found on the startup banner in the terminal, in the browser via the `<meta name="version" content={ KioskVersion } />` tag or browser console or by visiting `/about`.

**Your Kiosk installation**
- Docker
- Binary

**Your Kiosk ENV or config file**
Please make sure to remove any sensitive data (such as the Immich API key and URL) before sharing your configuration. To view a sanitized version of your config, visit `/config` while either the `KIOSK_DEBUG` or `KIOSK_DEBUG_VERBOSE` environment variables are set to true, or set `debug` or `debug_verbose` to true in your config.yaml file.

**Any parms passed to the URL used to access Kiosk**
- http://****/?show_time=true&time_format=12

**Desktop:**
 - OS: [e.g. iOS]
 - Browser [e.g. chrome, safari]
 - Version [e.g. 22]

**Smartphone:**
 - Device: [e.g. iPhone6]
 - OS: [e.g. iOS8.1]
 - Browser [e.g. stock browser, safari]
 - Version [e.g. 22]

**Additional context**
Add any other context about the problem here.

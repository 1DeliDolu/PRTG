# Getting started

## Project create

```
npx @grafana/create-plugin@latest
```

![1741512486202](image/DEVELOPMENT/1741512486202.png)

- Press the 'Y'.

![1741512527329](image/DEVELOPMENT/1741512527329.png)

- Select the data source with the down key.

![1741512547522](image/DEVELOPMENT/1741512547522.png)

- Enter the plugin name.

![1741512583576](image/DEVELOPMENT/1741512583576.png)

![1741512620482](image/DEVELOPMENT/1741512620482.png)

- Enter the Grafana organization name. If you haven't registered, proceed with the `Initial Steps` to create a cloud account first, then enter the name here and continue.

![1741512650284](image/DEVELOPMENT/1741512650284.png)

    *`cd ./grafana-prtg-datasource`
    * `npm install` to install frontend dependencies.
    * `npm exec playwright install chromium` to install e2e test dependencies.
    * `npm run dev` to build (and watch) the plugin frontend code.
    * `mage -v build:linux` to build the plugin backend code. Rerun this command every time you edit your backend files.
    * `docker compose up` to start a grafana development server.
    * Open http://localhost:3000 in your browser to create a dashboard to begin developing your plugin.

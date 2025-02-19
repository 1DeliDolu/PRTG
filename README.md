# PRTG Grafana Datasource Plugin

This repository contains a Grafana datasource plugin for PRTG, allowing users to visualize and analyze PRTG metrics within Grafana.

## Introduction

This Grafana datasource plugin integrates with PRTG, enabling users to fetch and display data from PRTG sensors directly in Grafana dashboards. It provides a seamless way to monitor and analyze PRTG data using Grafana's powerful visualization tools.

## Installation

First you need to download docker and go(golang) to your computer

Install [Docker](https://www.docker.com/)

Install [Go](https://go.dev/dl/)

1. Clone the repository:

   ```sh
   git clone https://github.com/1DeliDolu/PRTG.git
   ```
2. Navigate to the plugin directory:

   ```sh
   cd PRTG/maxmarkusprogram-prtg-datasource
   ```
3. Install dependencies:

   ```sh
   npm install
   ```
4. Build the plugin:

   ```sh
   npm run build
   ```
5. Copy the plugin to Grafana's plugin directory:

   ```sh
   cp -r dist /var/lib/grafana/plugins/PRTG
   ```
6. mage

   ```
   mage
   ```
7. Restart Grafana:

   ```sh
   sudo systemctl restart grafana-server
   ```

### **... or**

After clone, copy ***Prtg Folder*** to C:\Program Files\GrafanaLabs\grafana\data\plugins

**Restart Grafana:**

```sh
sudo systemctl restart grafana-server
```


## Configuration

1. Open Grafana and navigate to the Data Sources page.
2. Click on "Add data source" and select "PRTG".
3. Configure the PRTG datasource by providing the necessary connection details such as PRTG server URL, API key, and other relevant settings.
4. Save and test the datasource to ensure it is working correctly.

## Usage

1. Create a new dashboard or open an existing one in Grafana.
2. Add a new panel and select the PRTG datasource.
3. Configure the query to fetch data from the desired PRTG sensors.
4. Customize the visualization settings to display the data as needed.

## Troubleshooting

If you encounter any issues, please refer to the following troubleshooting steps:

- Ensure the PRTG server URL and API key are correctly configured.
- Check the Grafana server logs for any error messages.
- Verify that the plugin is correctly installed in Grafana's plugin directory.
- Restart Grafana and try again.

## Additional Resources

- [Grafana Plugin Development Documentation](https://grafana.com/developers/plugin-tools/)
- [PRTG API Documentation](https://www.paessler.com/manuals/prtg/api)

Feel free to contribute to this project by submitting issues or pull requests.

You can now save this content in the README.md file in your repository.

## Config Editor

1. Open `http://localhost:3000/` in browser or  `http://grafana.prtg:3000/connections/datasources`

![1739793462631](image/README/1739793462631.png)

2. Press PRTG

  ![1739793798353](image/README/1739793798353.png)

3. Enter your prtg server

![1739793866048](image/README/1739793866048.png)

4. Enter your api-token and Press save & test

  ![1739793921893](image/README/1739793921893.png)

5.Press Build an Dashboard

![1739794001603](image/README/1739794001603.png)

6.Press Add visualizaton[
]()
![1739794068185](image/README/1739794068185.png)

7.Press PRTG

![1739794166798](image/README/1739794166798.png)

## Query Metrics

Select Query Type

![1739795234405](image/README/1739795234405.png)

9.Select Group

![1739795274666](image/README/1739795274666.png)

10.Select Device

![1739795311709](image/README/1739795311709.png)

11.Select Sensor

![1739795351207](image/README/1739795351207.png)

12.Select Channel

![1739795402834](image/README/1739795402834.png)

13.Look at Panel

![1739795452206](image/README/1739795452206.png)

## Options

14.Add Group name in panel
![1739795578616](image/README/1739795578616.png)

15.Add Device and Sensor
![1739795687023](image/README/1739795687023.png)

16.Add new query
![1739795739188](image/README/1739795739188.png)

17.Select Query, Group, Device, Sensor and Channel
![1739795941422](image/README/1739795941422.png)

18.An another example
![1739796156994](image/README/1739796156994.png)

19.Fill Opacity
![1739796291106](image/README/1739796291106.png)

20.Select Stat
![1739796324396](image/README/1739796324396.png)

## Query Raw

1.Select Query Raw, Group, Device, Sensor,  Property and Filter Property

  ![1739796514348](image/README/1739796514348.png)

2. Examples

  ![1739796591456](image/README/1739796591456.png)

## Query Text

1.Query Text
![1739796808632](image/README/1739796808632.png)
![1739796830021](image/README/1739796830021.png)

## Panel

![1739797181230](image/README/1739797181230.png)

![1739797385328](image/README/1739797385328.png)

![1739797413883](image/README/1739797413883.png)

## Video

1. copy repository  https://github.com/1DeliDolu/PRTG.git

<video width="1000" height="500" controls>
  <source src="./video/clone.mp4" type="video/mp4">
</video>
``

2. open vs code and terminal with bash. Clone Repository

<video width="1000" height="500" controls>
  <source src="./video/vsc_clone.mp4" type="video/mp4">
</video>
``

3. Open new terminal with  wsl (i use debian) and  land "maxmarkusprogram-prtg-datasource" cd maxmarkusprogram-prtg-datasource

<video width="10000" height="500" controls>
  <source src="./video/cd.mp4" type="video/mp4">
</video>
``

4. npm install

<video width="1000" height="500" controls>
  <source src="./video/npminstall.mp4" type="video/mp4">
</video>
``

5. npm run build

<video width="1000" height="500" controls>
  <source src="./video/build.mp4" type="video/mp4">
</video>
``

7. mage

<video width="1000" height="500" controls>
  <source src="./video/mage.mp4" type="video/mp4">
</video>
``

7. mv dist/ Prtg

<video width="1000" height="500" controls>
  <source src="./video/prtg.mp4" type="video/mp4">
</video>
``

8. close vs code

<video width="1000" height="500" controls>
  <source src="./video/closevsc.mp4" type="video/mp4">
</video>
``

9. copy

<video width="1000" height="500" controls>
  <source src="./video/copy.mp4" type="video/mp4">
</video>
``

10. paste Prtg in C:\Program Files\GrafanaLabs\grafana\data\plugins

<video width="1000" height="500" controls>
  <source src="./video/paste.mp4" type="video/mp4">
</video>
``

11. Stop and start Grafana with powershell

<video width="1000" height="500" controls>
  <source src="./video/stop-start-grafana.mp4" type="video/mp4">
</video>
``

12. Sign in

    **at firts time
    `user admin pasword admin`**

<video width="1000" height="500" controls>
  <source src="./video/anmeldung.mp4" type="video/mp4">
</video>
``

13. Datasource panel

<video width="1000" height="500" controls>
  <source src="./video/datasource.mp4" type="video/mp4">
</video>
``

14.Create query

<video width="1000" height="500" controls>
  <source src="./video/query.mp4" type="video/mp4">
</video>
``

15. dashboard

<video width="1000" height="500" controls>
  <source src="./video/dashboard.mp4" type="video/mp4">
</video>
``

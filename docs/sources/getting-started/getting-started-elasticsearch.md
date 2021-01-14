+++
title = "With Grafana and Elasticsearch"
description = "Guide for getting started with Grafana"
keywords = ["grafana", "intro", "guide", "started", "Elastic", "Elasticsearch"]
aliases = ["/docs/grafana/latest/guides/gettingstarted","/docs/grafana/latest/guides/getting_started"]
weight = 400
+++

# Getting started with Grafana and Elasticsearch

Elasticsearch is a popular open source search and analytics engine for which Grafana provides out-of-the-box support. This topic walks you through the steps to create your first Elasticsearch backed dashboard in grafana to display any kind of data stored in your Elasticsearch database.

## Step 1. Install Grafana and build your first dashboard

Use the instructions in [Getting started with Grafana]({{< relref "getting-started.md" >}}) to:

- Install Grafana.
- Log in to Grafana.
- Create your first dashboard.

## Step 2. Download and install Elasticsearch

Elasticsearch can be installed on many different operating systems. Refer to the [Installing Elasticsearch](https://www.elastic.co/guide/en/elasticsearch/reference/current/install-elasticsearch.html) for a complete list of all available options.

Alternately you can install Elasticsearch using the resources available in [grafana/grafana](https://github.com/grafana/grafana) GitHub repository (recommended). Here you will find a collection of supported data sources, including Elasticsearch, along with test data and pre-configured dashboards for use.

1. Clone the [grafana/grafana](https://github.com/grafana/grafana/tree/master) repository to your local system.
1. Install Docker or verify that it is installed on your machine.
1. Within your local `grafana` repository, change directory to [devenv](https://github.com/grafana/grafana/tree/master/devenv).
   ```
   cd devenv
   ```
1. Run the bash command to setup data sources and dashboards.
   ```
   ./setup.sh
   ```
1. Restart the Grafana server.
1. Change directory back to the root directory.
   ```
   cd ..
   ```
1. Run the make command to create and start Elasticsearch.
   ```
    make devenv sources=elastic7
   ```

This will create and start Elasticsearch, filebeat, metricbeat, kibana and a process that sends random data to your new Elasticsearch instance. We'll ignore filebeat, metricbeat and kibana in this guide.

## Step 3. Configure and add the Elasticsearch data source

Once you have your Elasticsearch instance up and running and some process pushing data to it, you can configure your data source within Grafana.

To add the Elasticsearch data source:

1. In the Grafana side menu, hover your cursor over the **Configuration** (gear) icon and then click **Data Sources**.
1. Select the **Elasticsearch** option.
1. Click **Add data source** in the top right header to open the configuration page.
1. Enter the information specified in the tables below, then click **Save & Test**.

> **Note**: We'll assume you used the instructions from Grafana's devenv for simplicity, if this is not the case you may have to adapt the values in the tables below to match your use-case.

### General Settings

| Name   | Description                                                                           |
| ------ | ------------------------------------------------------------------------------------- |
| `Name` | The data source name. This is how you refer to the data source in panels and queries. |

### HTTP Settings

| Name     | Description                                                                                               | Value                    |
| -------- | --------------------------------------------------------------------------------------------------------- | ------------------------ |
| `URL`    | The IP address/hostname and optional port of your Elasticsearch instance.                                 | `http://localhost:12200` |
| `Access` | whether to connect to your Elasticsearch instance from the browser or through Grafana's backend (Server). | `Server (default)`       |

### Elasticsearch details

| Name              | Description                                                                | Value                  |
| ----------------- | -------------------------------------------------------------------------- | ---------------------- |
| `Index name`      | The index pattern used to identify indices you want to query with grafana. | `[metrics-]YYYY.MM.DD` |
| `Pattern`         | The time pattern for which the above index will be be interpolated.        | `Daily`                |
| `Time field name` | The name of the field that identiefies the timestamp of your documents.    | `@timestamp`           |
| `Version`         | The Elasticsearch version you are running.                                 | `7.0+`                 |

Once you click **Save & Test** the following message will appear, confirming that Grafana is able to communicate with your Elasticsearch instance:

![Elasticsearch confirmation message](/img/docs/getting-started/elasticsearch/confirmation-7-4.png)

## Step 4. Create your first Elasticsearch backed dashboard

1. In the Grafana side menu, hover your cursor over the **Create** (plus) icon and then click **Dashboard**.
1. Click the **Add new panel** button.

If you made Elasticsearch your default data source by toggling `default` on when creating it, it should be already selected:

![Elasticsearch panel edit](/img/docs/getting-started/elasticsearch/panel-edit-7-4.png)

Otherwise you can select it from the data source picker:

![data source picker](/img/docs/getting-started/elasticsearch/datasource-picker-7-4.png)

You can now start editing your query to get data out of your Elasticsearch instance.

## Next steps

Now that you have your Elasticsearch powered Grafana dashboard you can start exploring different ways of querying your Elasticsearch index to get the most out of your data.

You can read more about topics not covered in this guide, such as [Elasticsearch aggregations](https://www.elastic.co/guide/en/elasticsearch/reference/current/search-aggregations.html), [metricbeat](https://www.elastic.co/beats/metricbeat) and [filebeat](https://www.elastic.co/beats/filebeat) on [Elastic's website](https://www.elastic.co).

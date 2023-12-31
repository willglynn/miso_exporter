# `miso_exporter`

This is a Prometheus/OpenMetrics exporter for [MISO](https://www.misoenergy.org) data.

## Quickstart

Container images are available at [Docker Hub](https://hub.docker.com/r/willglynn/miso_exporter) and [GitHub
container registry](https://github.com/willglynn/purpleair_exporter/pkgs/container/miso_exporter).

```shell
$ docker run -it --rm -p 2023:2023 willglynn/miso_exporter
# or
$ docker run -it --rm -p 2023:2023 ghcr.io/willglynn/miso_exporter
2023/07/08 14:51:12 Starting HTTP server on :2023
```

## Endpoints

`GET /lmp` returns the Locational Marginal Price (LMP) for each `node` in each `region` of the MISO network. There are
four `kind`s of prices:

* `5m`, the realtime 5-minute price
* `1h`, the realtime hourly integrated price
* `exante`, the day-ahead ex ante price
* `expost`, the day-ahead ex post price

There are three `component`s:

* `LMP`, the full location marginal price
* `MLC`, the component of the price due to transmission losses
* `MCC`, the component of the price due to congestion

```text
# HELP miso_price_usd The price for power at a certain place and time
# TYPE miso_price_usd gauge
miso_price_usd{comp="LMP",kind="1h",node="AECI",region="Midwest"} 26.61 1688839200000
miso_price_usd{comp="LMP",kind="1h",node="AECI",region="Midwest"} 26.61 1688839260000
miso_price_usd{comp="LMP",kind="1h",node="AECI",region="Midwest"} 26.61 1688839320000
miso_price_usd{comp="LMP",kind="1h",node="AECI",region="Midwest"} 26.61 1688839380000
miso_price_usd{comp="LMP",kind="1h",node="AECI",region="Midwest"} 26.61 1688839440000
…
miso_price_usd{comp="LMP",kind="5min",node="AECI",region="Midwest"} 46.11 1688845200000
…
miso_price_usd{comp="LMP",kind="exante",node="AECI",region="Midwest"} 32.65 1688842800000
…
miso_price_usd{comp="LMP",kind="expost",node="AECI",region="Midwest"} 32.56 1688842800000
…
miso_price_usd{comp="MCC",kind="1h",node="AECI",region="Midwest"} -0.05 1688839200000
…
```

`GET /load` returns the total load, forecast and actual.

```text
# HELP miso_load_total_w The amount of load, forecast or actual
# TYPE miso_load_total_w gauge
miso_load_total_w{kind="actual"} 7.298e+10 1688792400000
miso_load_total_w{kind="actual"} 7.298e+10 1688792460000
miso_load_total_w{kind="actual"} 7.298e+10 1688792520000
miso_load_total_w{kind="actual"} 7.298e+10 1688792580000
miso_load_total_w{kind="actual"} 7.298e+10 1688792640000
miso_load_total_w{kind="actual"} 7.2789e+10 1688792700000
miso_load_total_w{kind="actual"} 7.2789e+10 1688792760000
miso_load_total_w{kind="actual"} 7.2789e+10 1688792820000
…
miso_load_total_w{kind="forecast"} 7.2009e+10 1688792400000
…
```

`GET /renewable_production` returns actual and forecast production from solar and wind.

```text
# HELP miso_renewable_production_w The amount of power produced from renewable sources
# TYPE miso_renewable_production_w gauge
miso_renewable_production_w{kind="actual",source="solar"} -5.02e+06 1688792400000
…
miso_renewable_production_w{kind="actual",source="solar"} 1.78845e+09 1688849940000
miso_renewable_production_w{kind="actual",source="wind"} 3.21474e+09 1688792400000
…
miso_renewable_production_w{kind="forecast",source="wind"} 7.952e+09 1688965140000
```

`GET /fuel` returns actual production by fuel type.

```text
# HELP miso_fuel_w The amount of power produced using a particular fuel
# TYPE miso_fuel_w gauge
miso_fuel_w{kind="actual",name="Coal"} 3.1159e+10 1688852400000
…
miso_fuel_w{kind="actual",name="Wind"} 4.604e+09 1688852640000
```

`GET /metrics` returns all of the above.

## Prometheus configuration

A good starting point:

```yaml
scrape_configs:
  - job_name: 'miso'
    scrape_interval: 2m
    scrape_timeout: 1m
    static_configs:
      - targets:
          - http://localhost:2023
    metric_relabel_configs:
      - if: '{__name__=~"^miso_.*"}'
        action: labeldrop
        regex: "instance|job"
```

`/metrics` returns unusually large responses, exceeding e.g. VictoriaMetrics' default maximum scrape size of 16 MB.
`-promscrape.maxScrapeSize=67108864` seems to work.

As of this writing, `miso_price_usd` (via `/metrics` or `/lmp`) produces 4116 series. This is very high cardinality but
it is also very low churn; series are created or destroyed only as the set of nodes changes, and MISO nodes correspond
to long-lived capital projects like substations or power plants.

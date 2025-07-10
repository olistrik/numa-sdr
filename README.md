# Numa SDR

> [!CAUTION] 
> This is a work in progress. Numa SDR was produced for during my
> three week internship at ASTRON as part of my honors programme. There are
> many features we would have liked to have added. Ultimately however, it was
> neither necessary nor was there time to implement them. As it stands, the
> most complete component is the `rtl_power` web monitor pipe. There is
> certainly room for improvement, but we found it very helpful for monitoring
> our SDR recordings remotely.

## Web Monitor

The web monitor can be spliced into an existing `rtl_power` pipeline and used
to produce a live plot of the active data stream. Currently, Numa does not
handle data storage so recovery of the complete graph is not possible.

```bash
# creates 160 bins across the FM band, individual stations should be visible
rtl_power -f 88M:108M:125k | numa_web > fm_stations.csv

# collect data for one hour and compress it on the fly
rtl_power -f ... -e 1h | numa_web | gzip > log.csv.gz
```

> [!NOTE]
> `numa_web` will not stop when `rtl_power` exits. If your downstream tool
> expects or requires that, you may experience issues.


### Arguments

`numa_web` support a number of arguments:

```
--address ip        -a ip       The IP address to listen on. Defaults to '0.0.0.0'.
--port int          -p int      The Port to listen on. Defaults to '21753'.
--offset float      -o float    The frequency offset when using an up/down converter.
--history duration              The maximum timespan of data to cache for new connections. Defaults to 1h'. 
--title string,     -t string   The title of the webpage. Defaults to 'Numa'.
--help              -h          Display the help text.
```

Notibly, the history flag can be used to control the size of the in-memory
cache maintained by numa for new connections. This is not the same as storage
and should not be set too high. Depending on the datarate of your scan this can
rapidly consume large amounts of memory. Additionally, new connections to the
website will be sent the entire cache, regardless of how many rows they can
actually display.

### Building

Numa is written in Golang. So you sould first install and configure go. [The
official install instructions can be found here.](https://go.dev/doc/install)
If using ubuntu do not use `apt`.

> [!NOTE]
> `numa_web` does not _yet_ depend on the `rtl_sdr` or `soapy` libraries, but 
> they may become necessary in the future.


To build `numa_web` for your current system run:

```bash
go build -o numa_web ./cmd/web
```

It is also possible to cross-compile numa for other architectures, for example
to build for an `aarm64` system such as a Raspberry Pi:

```bash
GOARCH=arm64 GOOS=linux go build -o ./cmd/web
```

### Future Work

These are some ideas that we would like to implement in the future.

- [ ] Support storing data in a PostgreSQL database.
    - [ ] Allow the frontend to retrieve arbirary timespans of historical data.
- [ ] Replace plotly.js with something more optimal for plotting large dynamic
      waterfall plots.
- [ ] Implement `numa_power` using soapy to support more SDRs.
    - [ ] Allow for custom processing pipelines _before_ integration (maybe via GNURadio?).



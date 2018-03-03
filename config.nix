{
    Logger = {
        Formatter = "json";
        Level = "info";
    };
    HTTP = {
        Addr = "127.0.0.1:8080";
    };
    Handlers = [
        {
            Method = "get";
            Path = "/api/v1/tickers";
            Type = "latest";
            Latest = {
                Format = "json";
                Consumer = {
                    Format = "json";
                    Queue = {
                        Type = "nsq";
                        Nsq = {
                            Addr = "127.0.0.1:4150";
                            Topic = "tickers";
                            Channel = "platypus-latest";
                            ConsumerBufferSize = 128;
                            LogLevel = "info";
                        };
                    };
                };
                Store = {
                    Type = "memoryttl";
                    MemoryTTL = {
                        TTL = "24h";
			                  Resolution = "1m";
		                };
                };
                Key = ''{{.market}}|{{(index .currencyPair 0).symbol}}|{{(index .currencyPair 1).symbol}}'';
            };
        }
        {
            Method = "get";
            Path = "/api/v1/tickers-changes";
            Type = "latests";
            Latests = {
                Format = "json";
                Inputs = [
                    {
                        Consumer = {
                            Format = "json";
                            Queue = {
                                Type = "nsq";
                                Nsq = {
                                    Addr = "127.0.0.1:4150";
                                    Topic = "tickers";
                                    Channel = "platypus-tc-latests";
                                    ConsumerBufferSize = 128;
                                };
                            };
                        };
                        Store = {
                            Type = "memoryttl";
                            MemoryTTL = {
                                TTL = "24h";
				                        Resolution = "1m";
			                      };
			                  };
                        Key = ''{{.market}}|{{(index .currencyPair 0).symbol}}|{{(index .currencyPair 1).symbol}}'';
                    }
                    {
                        Consumer = {
                            Format = "json";
                            Queue = {
                                Type = "nsq";
                                Nsq = {
                                    Addr = "127.0.0.1:4150";
                                    Topic = "changes";
                                    Channel = "platypus-tc-latests";
                                    ConsumerBufferSize = 128;
                                };
                            };
                        };
                        Store = {
                            Type = "memoryttl";
                            MemoryTTL = {
                                TTL = "24h";
				                        Resolution = "1m";
			                      };
                        };
                        Key = ''{{.period}}|{{(index .currencyPair 0).symbol}}|{{(index .currencyPair 1).symbol}}'';
                    }
                ];
                Wrap = ''{{"{"}}{{range $i, $e := $}}{{if $i}},{{end}}"{{- $e.Input.Consumer.Queue.Nsq.Topic -}}":{{- (printf "%s" $e.Events.JSON) -}}{{end}}{{"}"}}'';
            };
        }
        {
            Method = "get";
            Path = "/api/v1/tickers/stream";
            Type = "stream";
            Stream = {
                Format = "json";
                Writer = {
                    ScheduleTimeout = "10s";
                    Pool = {
                        Workers = 128;
                        QueueSize = 1024;
                    };
                };
                Consumer = {
                    Format = "json";
                    Queue = {
                        Type = "nsq";
                        Nsq = {
                            Addr = "127.0.0.1:4150";
                            Topic = "tickers";
                            Channel = "platypus-stream";
                            ConsumerBufferSize = 128;
                            LogLevel = "info";
                        };
                    };
                };
            };
        }
        {
            Method = "get";
            Path = "/api/v1/tickers-changes/stream";
            Type = "streams";
            Streams = {
                Format = "json";
                Inputs = [
                    {
                        Consumer = {
                            Format = "json";
                            Queue = {
                                Type = "nsq";
                                Nsq = {
                                    Addr = "127.0.0.1:4150";
                                    Topic = "tickers";
                                    Channel = "platypus-tc-streams";
                                    ConsumerBufferSize = 128;
                                };
                            };
                        };
                    }
                    {
                        Consumer = {
                            Format = "json";
                            Queue = {
                                Type = "nsq";
                                Nsq = {
                                    Addr = "127.0.0.1:4150";
                                    Topic = "changes";
                                    Channel = "platypus-tc-streams";
                                    ConsumerBufferSize = 128;
                                };
                            };
                        };
                    }
                ];
                Wrap = ''{"type":"{{- .Input.Consumer.Queue.Nsq.Topic -}}","payload":{{- (printf "%s" .Event.JSON) -}}}'';
                Writer = {
                    ScheduleTimeout = "10s";
                    Pool = {
                        Workers = 128;
                        QueueSize = 1024;
                    };
                };
            };
        }
    ];
}
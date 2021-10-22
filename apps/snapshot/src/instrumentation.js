import {AlwaysOnSampler, ParentBasedSampler, W3CTraceContextPropagator} from "@opentelemetry/core";
import {WebTracerProvider} from "@opentelemetry/sdk-trace-web";
import {Resource} from "@opentelemetry/resources";
import {BatchSpanProcessor, ConsoleSpanExporter, SimpleSpanProcessor} from "@opentelemetry/sdk-trace-base";
import {OTLPTraceExporter} from "@opentelemetry/exporter-otlp-http";
import {ZoneContextManager} from "@opentelemetry/context-zone";
import {registerInstrumentations} from "@opentelemetry/instrumentation";
import * as autoInstrumentationAPI from "@opentelemetry/auto-instrumentations-web";
import * as api from "@opentelemetry/api";
import {SemanticResourceAttributes} from "@opentelemetry/semantic-conventions";


export function setupInstrumentation() {
    /* Set Global Propagator */
    // api.propagation.setGlobalPropagator(new W3CTraceContextPropagator());
    console.log("no propagator")
    const provider = new WebTracerProvider({
        resource: new Resource({
            // https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/semantic_conventions/README.md#semantic-attributes-with-sdk-provided-default-value
            [SemanticResourceAttributes.SERVICE_NAME]: "snapshot",
            [SemanticResourceAttributes.SERVICE_NAMESPACE]: "example.meetup",
            [SemanticResourceAttributes.SERVICE_VERSION]: "0.1.0",
        }),
        // this is the same as directly using the AlwaysOnSampler, but this shows how to respect the parent.
        // https://github.com/open-telemetry/opentelemetry-js/tree/main/packages/opentelemetry-core#parentbased-sampler
        sampler: new ParentBasedSampler({
            // By default, the ParentBasedSampler will respect the parent span's sampling
            // decision. This is configurable by providing a different sampler to use
            // based on the situation. See configuration details above.
            //
            // This will delegate the sampling decision of all root traces (no parent)
            // to the TraceIdRatioBasedSampler.
            // See details of TraceIdRatioBasedSampler above.
            root: new AlwaysOnSampler()
        }),
    });

    // add processors
    provider.addSpanProcessor(new SimpleSpanProcessor(new ConsoleSpanExporter()));

    const exporter = new OTLPTraceExporter({
        url: "http://collector.testing.com/v1/traces", // url is optional and can be omitted - default is http://localhost:55681/v1/traces
        headers: {}, // an optional object containing custom headers to be sent with each request
        concurrencyLimit: 10, // an optional limit on pending requests
    });
    provider.addSpanProcessor(new BatchSpanProcessor(exporter, {
        // The maximum queue size. After the size is reached spans are dropped.
        maxQueueSize: 100,
        // The maximum batch size of every export. It must be smaller or equal to maxQueueSize.
        maxExportBatchSize: 10,
        // The interval between two consecutive exports
        scheduledDelayMillis: 500,
        // How long the export can run before it is cancelled
        exportTimeoutMillis: 30000,
    }));




    provider.register({
        // Changing default contextManager to use ZoneContextManager - supports asynchronous operations - optional
        contextManager: new ZoneContextManager(),
        // use the W3C standard for trace propagation
        // propagator: new W3CTraceContextPropagator(),
    });

    registerInstrumentations({
        instrumentations: [
            autoInstrumentationAPI.getWebAutoInstrumentations({
                // load custom configuration for xml-http-request instrumentation
                '@opentelemetry/instrumentation-xml-http-request': {
                    clearTimingResources: true,
                },
            }),
        ],
    });

}
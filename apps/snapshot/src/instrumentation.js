import {
    AlwaysOnSampler,
    ParentBasedSampler,
    CompositePropagator,
    W3CTraceContextPropagator,
    W3CBaggagePropagator
} from "@opentelemetry/core";
import React from "react";
import {JaegerPropagator} from "@opentelemetry/propagator-jaeger";
import {WebTracerProvider} from "@opentelemetry/sdk-trace-web";
import {Resource} from "@opentelemetry/resources";
import {BatchSpanProcessor, ConsoleSpanExporter, SimpleSpanProcessor} from "@opentelemetry/sdk-trace-base";
import {OTLPTraceExporter} from "@opentelemetry/exporter-otlp-http";
import {ZoneContextManager} from "@opentelemetry/context-zone";
import {registerInstrumentations} from "@opentelemetry/instrumentation";
import * as autoInstrumentationAPI from "@opentelemetry/auto-instrumentations-web";
import * as api from "@opentelemetry/api";
import {BaseOpenTelemetryComponent} from '@opentelemetry/plugin-react-load';
import {SemanticResourceAttributes} from "@opentelemetry/semantic-conventions";
import {getUnleash} from "./unleash";
import {DiagConsoleLogger} from "@opentelemetry/api";


/* Set Global Propagator */
// support jaeger, W3C and W3C Baggage propagation
const propagator = new CompositePropagator({propagators: [new JaegerPropagator(), new W3CTraceContextPropagator(), new W3CBaggagePropagator()]})
api.propagation.setGlobalPropagator(propagator);
export const traceProvider = new WebTracerProvider({
    resource: new Resource({
        // https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/semantic_conventions/README.md#semantic-attributes-with-sdk-provided-default-value
        [SemanticResourceAttributes.SERVICE_NAME]: "snapshot",
        [SemanticResourceAttributes.SERVICE_NAMESPACE]: "example.meetup",
        [SemanticResourceAttributes.SERVICE_VERSION]: "0.1.0",
        [SemanticResourceAttributes.PROCESS_OWNER]: "1233"
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
traceProvider.addSpanProcessor(new SimpleSpanProcessor(new ConsoleSpanExporter()));

getUnleash().then(value => {
    const exporter = new OTLPTraceExporter({
        url: "http://collector.testing.com/v1/traces", // url is optional and can be omitted - default is http://localhost:55681/v1/traces
        headers: {}, // an optional object containing custom headers to be sent with each request
        concurrencyLimit: 10, // an optional limit on pending requests
    });

    const enabled = value.isEnabled("opentelemetry")
    console.log(`opentelemetry feature flag: ${enabled}`)

    if (enabled) {
        traceProvider.addSpanProcessor(new BatchSpanProcessor(exporter, {
            // The maximum queue size. After the size is reached spans are dropped.
            maxQueueSize: 100,
            // The maximum batch size of every export. It must be smaller or equal to maxQueueSize.
            maxExportBatchSize: 20,
            // The interval between two consecutive exports
            scheduledDelayMillis: 500,
            // How long the export can run before it is cancelled
            exportTimeoutMillis: 30000,
        }));
    }
}).catch(reason => console.log(reason))


traceProvider.register({
    // Changing default contextManager to use ZoneContextManager - supports asynchronous operations - optional
    contextManager: new ZoneContextManager(),
    // use the composite propagator for trace propagation
    propagator: propagator,
});

registerInstrumentations({
    instrumentations: [
        autoInstrumentationAPI.getWebAutoInstrumentations({
            // load custom configuration for xml-http-request instrumentation
            '@opentelemetry/instrumentation-xml-http-request': {
                clearTimingResources: true,
                propagateTraceHeaderCorsUrls: RegExp("/*.\\.testing\\.com/"), // send propagation header to this url(s)
            },
        }),
    ],

});

BaseOpenTelemetryComponent.setTracer("snapshot") // set global tracer for react
BaseOpenTelemetryComponent.setLogger(new DiagConsoleLogger())


export default function TraceProvider({children}) {
    return (
        <>
            {children}
        </>
    );
}
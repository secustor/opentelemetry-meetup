import React, { createContext, useState } from "react";
import axios from "axios";
import { apiKey } from "../api/config";
import {traceProvider} from "../instrumentation";
import {SpanStatusCode} from "@opentelemetry/api";

export const PhotoContext = createContext();

const PhotoContextProvider = props => {
  const [images, setImages] = useState([]);
  const [loading, setLoading] = useState(true);

  const runSearchTracer = traceProvider.getTracer('snapshot')


  const runSearch = query => {
    runSearchTracer.startActiveSpan("search images",span => {
      runSearchTracer.startActiveSpan("query images",{}, span => {  // https://github.com/open-telemetry/opentelemetry-js/issues/1923
        axios
            .get(
                `https://api.flickr.com/services/rest/?method=flickr.photos.search&api_key=${apiKey}&tags=${query}&per_page=24&format=json&nojsoncallback=1`
            )

            .then(response => {
              setImages(response.data.photos.photo);
              setLoading(false);
              span.setStatus({ code: SpanStatusCode.OK });
            })

            .catch(error => {
              console.log(
                  "Encountered an error with fetching and parsing data",
                  error
              );
              span.recordException(error)
              span.setStatus({
                code: SpanStatusCode.ERROR,
                message: error.message
              })
            }).finally(() => span.end());
      })

      runSearchTracer.startActiveSpan("report search", span => {
        axios.post(`http://report.testing.com/kafka/receiver`,query).then(response => {
          console.log(`send data ${query} and got status: ${response.status}`)
          span.setStatus({ code: SpanStatusCode.OK });
        }).catch(error => {
          console.log(
              "Encountered an error with sending data",
              error
          );
          span.recordException(error)
          span.setStatus({
            code: SpanStatusCode.ERROR,
            message: error.message
          })
        }).finally(() => span.end())
      })
      span.end()
    })
  };

  return (
    <PhotoContext.Provider value={{ images, loading, runSearch }}>
      {props.children}
    </PhotoContext.Provider>
  );
};

export default PhotoContextProvider;

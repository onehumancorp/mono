import React from "react";
import ReactDOM from "react-dom/client";
import { App } from "./App";
import "./styles.css";

/**
 * @summary Entry point for the React application.
 * @param None
 * @returns None
 * @throws None
 * @remarks Side Effects: Mounts the application to the DOM.
 */
ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);

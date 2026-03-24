import React from "react";
import ReactDOM from "react-dom/client";
import { App } from "./App";
import "./styles.css";

/**
 * Summary: Entry point for the React application.
 * Intent: Entry point for the React application.
 * Params: None
 * Returns: None
 * Errors: None
 * Side Effects: Mounts the application to the DOM.
 */
ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);

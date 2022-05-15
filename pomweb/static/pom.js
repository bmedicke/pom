"use strict";

const ws = new WebSocket("ws://192.168.8.171:8421/ws");
ws.addEventListener("message", (event) => {
  const payload = JSON.parse(event.data);
  var seconds = payload.Remaining / 1e9;
  const minutes = Math.floor(seconds / 60);
  seconds = Math.floor(seconds - minutes * 60);
  seconds = String(seconds).padStart(2, "0");
  document.title = `${minutes}:${seconds} ${payload.State}`;
});

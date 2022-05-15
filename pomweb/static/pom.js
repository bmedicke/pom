'use strict'

function timer(props) {
  return <h1>Hello world</h1>
}

const domContainer = document.querySelector("#timer_container")
const root = ReactDOM.createRoot(domContainer)
root.render(React.createElement(timer))

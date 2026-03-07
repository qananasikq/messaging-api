export function createDialogSocket(token: string, dialogId: string) {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:"
  const url = new URL(`${protocol}//localhost:8080/ws`)
  url.searchParams.set("token", token)

  const socket = new WebSocket(url)

  socket.addEventListener("open", () => {
    socket.send(JSON.stringify({ dialog_id: dialogId }))
  })

  return socket
}

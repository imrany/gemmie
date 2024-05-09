import { useState } from "react";
import Sidebar from "./Components/Sidebar/Sidebar";
import Chat from "./Components/Chat/Chat";

function App() {
  const [count, setCount] = useState(0);

  return (
    <>
      <Sidebar />
      <Chat />
    </>
  );
}

export default App;

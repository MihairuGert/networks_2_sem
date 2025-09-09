package sender;

import java.io.*;
import java.net.ServerSocket;
import java.net.Socket;
import java.util.concurrent.ConcurrentHashMap;

public class Server {
    private final int port;
    private String relativeDir = "/uploads";
    private final ConcurrentHashMap<ClientData, Integer> clients;
    private final ServerSocket serverSocket;

    public Server(int port) throws Exception {
        if (port > Short.MAX_VALUE*2 - 1) {
            throw new Exception("invalid port");
        }
        this.port = port;
        clients = new ConcurrentHashMap<>();
        serverSocket = new ServerSocket(port);
    }

    public void startListen() {
        try {
            Socket socket = serverSocket.accept();
            ClientData clientData = new ClientData(socket);
            clients.put(clientData, 1);
            new Thread(this::startListen).start();
            initializeReceive(clientData);
        } catch (IOException e) {
            System.out.println(e.getMessage());
        }
    }

    private void initializeReceive(ClientData clientData) {
        BufferedReader in = null;
        try {
            in = new BufferedReader(new InputStreamReader(clientData.getSocket().getInputStream()));
            BufferedWriter out = new BufferedWriter(new OutputStreamWriter(clientData.getSocket().getOutputStream()));
            String word = in.readLine();
            System.out.println(word);
            out.write("Привет, это Сервер! Подтверждаю, вы написали : " + word + "\n");
            out.flush();
        } catch (IOException e) {
            System.out.println(e.getMessage());
        }
    }

    public static void main(String[] args) {
        Server server = null;
        try {
            server = new Server(8000);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
        server.startListen();
    }
}

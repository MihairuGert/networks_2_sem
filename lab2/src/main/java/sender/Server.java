package sender;

import java.net.ServerSocket;
import java.util.concurrent.ConcurrentHashMap;

public class Server {
    private final int port;
    private String relativeDir = "/uploads";
    private final ConcurrentHashMap<ClientData, Integer> clients;
    private final ServerSocket serverSocket;

    Server(int port) throws Exception {
        if (port > Short.MAX_VALUE*2 - 1) {
            throw new Exception("invalid port");
        }
        this.port = port;
        clients = new ConcurrentHashMap<>();
        serverSocket = new ServerSocket(port);
    }

    public static void main(String[] args) {

    }
}

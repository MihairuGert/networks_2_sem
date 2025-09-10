package sender;

import java.io.*;
import java.net.ServerSocket;
import java.net.Socket;
import java.util.concurrent.ConcurrentHashMap;

public class Server {
    private final int port;
    private String relativeDir = "./uploads";
    private final ConcurrentHashMap<ClientData, Integer> clients;
    private final ServerSocket serverSocket;
    private BufferedReader in;
    private BufferedWriter out;

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
            receive(clientData);
        } catch (IOException e) {
            System.out.println(e.getMessage());
        }
    }

    private String getFilename() throws Exception {
        try {
            String str = in.readLine();
            if (str.matches("FILENAME=.+")) {
                return str.split("=")[1];
            } else {
                throw new Exception("wrong protocol");
            }
        } catch (IOException e) {
            throw new Exception(e);
        }
    }

    private BufferedWriter createFile(String filename) {
        try {
            File directory = new File(relativeDir);
            if (!directory.exists()) {
                if (!directory.mkdirs())
                    throw new IOException("Failed to create directory: " + relativeDir);
            }
            return new BufferedWriter(new FileWriter(relativeDir + File.separator + filename));
        } catch (IOException e) {
            e.printStackTrace();
            throw new RuntimeException("Failed to create file: " + filename, e);
        }
    }

    private void getFile(String filename) {
        try(BufferedWriter bufferedWriter = createFile(filename)) {
            final int rawDataSize = 512;
            char[] rawData = new char[rawDataSize];
            while(true) {
                try {
                    int symRead = in.read(rawData, 0, 512);
                    if (symRead == -1) {
                        break;
                    }
                    bufferedWriter.write(rawData, 0, symRead);
                } catch (IOException e) {
                    throw new RuntimeException(e);
                }
            }
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

    private void receive(ClientData clientData) {
        String filename;
        try {
            in = new BufferedReader(new InputStreamReader(clientData.getSocket().getInputStream()));
            out = new BufferedWriter(new OutputStreamWriter(clientData.getSocket().getOutputStream()));

            filename = getFilename();
            System.out.println(filename);

            getFile(filename);

            in.close();
            out.close();
            clientData.getSocket().close();
        } catch (Exception e) {
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

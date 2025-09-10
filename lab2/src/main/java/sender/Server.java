package sender;

import java.io.*;
import java.net.ServerSocket;
import java.net.Socket;
import java.util.Arrays;
import java.util.concurrent.ConcurrentHashMap;

public class Server {
    private final int port;
    private String relativeDir = "./uploads";
    private final ConcurrentHashMap<ClientData, Integer> clients;
    private final ServerSocket serverSocket;
    private FileInputStream in;
    private FileOutputStream out;

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
            byte[] bytes = new byte[4096];
            int bytesRead = in.read(bytes);
            String msg = Arrays.toString(bytes);
            if (msg.matches("FILENAME=.+")) {
                return msg.split("=")[1];
            } else {
                throw new Exception("wrong protocol");
            }
        } catch (IOException e) {
            throw new Exception(e);
        }
    }

    private FileOutputStream createFile(String filename) {
        try {
            File directory = new File(relativeDir);
            if (!directory.exists()) {
                if (!directory.mkdirs())
                    throw new IOException("Failed to create directory: " + relativeDir);
            }
            return new FileOutputStream(relativeDir + File.separator + filename);
        } catch (IOException e) {
            e.printStackTrace();
            throw new RuntimeException("Failed to create file: " + filename, e);
        }
    }

    private void getFile(String filename) {
        try(FileOutputStream fileOutputStream = createFile(filename)) {
            final int rawDataSize = 512*4;
            byte[] rawData = new byte[rawDataSize];
            while(true) {
                try {
                    int symRead = in.read(rawData, 0, rawDataSize);

                    if (symRead == -1) {
                        break;
                    }
                    fileOutputStream.write(rawData, 0, symRead);
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
            in = (FileInputStream) clientData.getSocket().getInputStream();
            out = (FileOutputStream) clientData.getSocket().getOutputStream();

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

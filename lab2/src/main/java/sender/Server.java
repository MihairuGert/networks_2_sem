package sender;

import java.io.*;
import java.net.ServerSocket;
import java.net.Socket;
import java.text.DecimalFormat;
import java.util.Arrays;
import java.util.Iterator;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

public class Server {
    private final int port;
    private String relativeDir = "./uploads";

    private final ConcurrentHashMap<ClientData, ClientStatistics> clients;
    private final ServerSocket serverSocket;
    private InputStream in;
    private OutputStream out;

    private final ScheduledExecutorService scheduler = Executors.newScheduledThreadPool(1);

    private void printClientsInfo() {
        Iterator<Map.Entry<ClientData, ClientStatistics>> iterator = clients.entrySet().iterator();
        int count = 0;
        DecimalFormat decimalFormat = new DecimalFormat("0.00");
        while (iterator.hasNext()) {
            Map.Entry<ClientData, ClientStatistics> entry = iterator.next();
            count++;
            System.out.println("<--Client #" + count + "-->");
            System.out.println("   Filename: " + entry.getValue().getFilename());
            System.out.println("   Instant Speed: " + decimalFormat.format(entry.getValue().getInstantSpeed()) + " Kb/s");
            entry.getValue().setBytesReceivedPeriodAgo();
            System.out.println("   Average Speed: " + decimalFormat.format(entry.getValue().getAverageSpeed()) + " Kb/s");
            System.out.println();
        }
    }

    public Server(int port) throws Exception {
        if (port > Short.MAX_VALUE*2 - 1) {
            throw new Exception("invalid port");
        }
        this.port = port;
        clients = new ConcurrentHashMap<>();
        serverSocket = new ServerSocket(port);
        scheduler.scheduleAtFixedRate(this::printClientsInfo, 1, 3, TimeUnit.SECONDS);
    }

    public void startListen() {
        try {
            Socket socket = serverSocket.accept();
            ClientData clientData = new ClientData(socket);
            clients.put(clientData, new ClientStatistics(3));
            new Thread(this::startListen).start();
            receive(clientData, socket);
        } catch (IOException e) {
            System.out.println(e.getMessage());
        }
    }

    private String getFilename() throws Exception {
        try {
            byte[] bytes = new byte[4096];
            int bytesRead = in.read(bytes);
            String msg = new String(bytes).trim();
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

    private void getFile(ClientData clientData, String filename) throws Exception {
        try(FileOutputStream fileOutputStream = createFile(filename)) {
            final int rawDataSize = 512*4;
            byte[] rawData = new byte[rawDataSize];
            while(true) {
                try {
                    int symRead = in.read(rawData, 0, rawDataSize);
                    if (symRead == -1) {
                        break;
                    }
                    clients.get(clientData).addBytesReceived(symRead);
                    fileOutputStream.write(rawData, 0, symRead);
                } catch (IOException e) {
                    throw new Exception(e);
                }
            }
        } catch (IOException e) {
            throw new Exception(e);
        }
    }

    private void receive(ClientData clientData, Socket socket) throws IOException {
        String filename;
        try (socket) {
            in = clientData.getSocket().getInputStream();
            out = clientData.getSocket().getOutputStream();

            filename = getFilename();
            clients.get(clientData).setFilename(filename);
            System.out.println(filename);

            getFile(clientData, filename);

        } catch (Exception e) {
            System.out.println(e.getMessage());
        } finally {
            clients.remove(clientData);
            in.close();
            out.close();
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

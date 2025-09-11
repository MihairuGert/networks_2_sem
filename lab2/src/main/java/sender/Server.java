package sender;

import java.io.*;
import java.net.ServerSocket;
import java.net.Socket;
import java.sql.Time;
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

    private void printClientInfo(ClientData clientData) {
        ClientStatistics clientStatistics = clients.get(clientData);
        printStats(clientStatistics);
    }

    private static void printStats(ClientStatistics clientStatistics) {
        DecimalFormat decimalFormat = new DecimalFormat("0.00");
        System.out.println("   Filename: " + clientStatistics.getFilename());
        System.out.println("   Instant Speed: " + decimalFormat.format(clientStatistics.getInstantSpeed()) + " Mb/s");
        clientStatistics.setBytesReceivedPeriodAgo();
        System.out.println("   Average Speed: " + decimalFormat.format(clientStatistics.getAverageSpeed()) + " Mb/s");
        System.out.println("   Progress: " + decimalFormat.format(clientStatistics.getPercent()) + " %");
        System.out.println();
    }

    private void printClientsInfo() {
        Iterator<Map.Entry<ClientData, ClientStatistics>> iterator = clients.entrySet().iterator();
        int count = 0;
        while (iterator.hasNext()) {
            Map.Entry<ClientData, ClientStatistics> entry = iterator.next();
            count++;
            System.out.println("<--Client #" + count + "-->");
            printStats(entry.getValue());
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
            String msg = getStringMsg();
            if (msg.matches("FILENAME=.+")) {
                return msg.split("=")[1];
            } else {
                throw new Exception("wrong protocol(filename)");
            }
        } catch (IOException e) {
            throw new Exception(e);
        }
    }

    private long getFileSize() throws Exception {
        try {
            String msg = getStringMsg();
            if (msg.matches("FILE_SIZE=.+")) {
                return Long.parseLong(msg.split("=")[1]);
            } else {
                throw new Exception("wrong protocol(file size)");
            }
        } catch (IOException e) {
            throw new Exception(e);
        }
    }

    private String getStringMsg() throws IOException {
        StringBuilder sb = new StringBuilder();
        int byteRead;

        while ((byteRead = in.read()) != -1) {
            char c = (char) byteRead;
            if (c == '\n') {
                break;
            }
            sb.append(c);
        }

        return sb.toString();
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
        try (socket) {
            long time_start = System.currentTimeMillis();

            in = clientData.getSocket().getInputStream();
            out = clientData.getSocket().getOutputStream();

            String filename = getFilename();
            clients.get(clientData).setFilename(filename);

            long file_size = getFileSize();
            clients.get(clientData).setFileSize(file_size);

            getFile(clientData, filename);

            long time_end = System.currentTimeMillis() - time_start;
            if (TimeUnit.MILLISECONDS.toSeconds(time_end) <= 3)
                printClientInfo(clientData);

            if (clients.get(clientData).getBytesReceived() >= file_size) {
                out.write("SUCCESS".getBytes());
            } else {
                out.write("FAILURE".getBytes());
            }
            out.flush();

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

package sender;

import java.io.IOException;
import java.net.InetAddress;
import java.net.Socket;
import java.io.*;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;

public class Client {

    private final String path;
    private final String filename;
    private Socket socket;
    private InputStream in;
    private OutputStream out;

    public Client(String path) {
        this.path = path;
        String[] spl = path.split("/");
        filename = spl[spl.length - 1];
        if (filename.length() * 16 > 4096) {
            throw new RuntimeException("wrong length");
        }
    }

    private void connect(String ip, int port) {
        try {
            socket = new Socket(InetAddress.getByName(ip), port);
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

    private void sendFilename() {
        try {
            out.write(("FILENAME="+filename+'\n').getBytes(StandardCharsets.UTF_8));
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

    private void sendFileSize(long size) {
        try {
            out.write(("FILE_SIZE="+size+'\n').getBytes());
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

    private FileInputStream getFileReader() {
        try {
            return new FileInputStream(path);
        } catch (IOException e) {
            e.printStackTrace();
            throw new RuntimeException(e);
        }
    }

    private void sendFile() {
        try (FileInputStream fileInputStream = getFileReader()) {
            final int rawDataSize = 512;
            byte[] rawData = new byte[rawDataSize];
            while(true) {
                try {
                    int symRead = fileInputStream.read(rawData, 0, 512);
                    if (symRead == -1) {
                        break;
                    }
                    out.write(rawData, 0, symRead);
                    out.flush();
                } catch (IOException e) {
                    throw new RuntimeException(e);
                }
            }
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

    private static long getFileSize(String filePath) throws IOException {
        Path path = Paths.get(filePath);
        return Files.size(path);
    }

    private boolean isSuccess() throws IOException {
        byte[] data = new byte[128];
        int bytesRead = in.read(data);
        String res = new String(data).trim();
        return res.equals("SUCCESS");
    }

    public void startSend(String ip, int port) throws IOException {
        connect(ip, port);
        try {
            in = socket.getInputStream();
            out = socket.getOutputStream();

            sendFilename();
            long file_size = getFileSize(path);
            if (file_size > 1024 * 1024 * 1024) {
                throw new IOException("File size must not exceed 1 Tb.");
            }
            sendFileSize(file_size);
            sendFile();

            socket.shutdownOutput();
            boolean isSuccess = isSuccess();
            printSendResult(isSuccess);

        } catch (IOException e) {
            System.out.println(e.getMessage());
        } finally {
            in.close();
            out.close();
            socket.close();
        }
    }

    private static void printSendResult(boolean isSuccess) {
        if (isSuccess) {
            System.out.println("File sending was completed successfully.");
        } else {
            System.out.println("File sending was failed.");
        }
    }

    public static void main(String[] args) {
        Client client = new Client("./yoy.jpg");
        try {
            client.startSend("192.168.0.120", 8000);
        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}

package sender;

import java.io.IOException;
import java.net.InetAddress;
import java.net.ServerSocket;
import java.net.Socket;
import java.net.SocketAddress;
import java.io.*;
import java.util.Scanner;

public class Client {

    private final String path;
    private final String filename;
    private Socket socket;
    private BufferedReader in;
    private BufferedWriter out;

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
            out.write("FILENAME="+filename+'\n');
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

    public void startSend(String ip, int port) {
        connect(ip, port);
        try {
            in = new BufferedReader(new InputStreamReader(socket.getInputStream()));
            out = new BufferedWriter(new OutputStreamWriter(socket.getOutputStream()));

            sendFilename();
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

    public static void main(String[] args) {
        Client client = new Client("hui");
        client.startSend("192.168.0.120", 8000);
    }
}

package sender;

import java.net.InetAddress;
import java.net.Socket;

public class ClientData {
    private String filename;
    private Socket socket;

    ClientData(Socket socket) {
        this.socket = socket;
    }

    public void setSocket(Socket socket) {
        this.socket = socket;
    }

    public Socket getSocket() {
        return socket;
    }

    public String getFilename() {
        return filename;
    }
}

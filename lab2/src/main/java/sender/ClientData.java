package sender;

import java.net.InetAddress;

public class ClientData {
    private final InetAddress address;
    private final String filename;

    ClientData(InetAddress address, String filename) {
        this.address = address;
        this.filename = filename;
    }

    public InetAddress getAddress() {
        return address;
    }

    public String getFilename() {
        return filename;
    }
}

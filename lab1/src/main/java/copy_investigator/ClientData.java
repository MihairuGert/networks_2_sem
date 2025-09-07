package copy_investigator;

import java.net.InetAddress;

public class ClientData {
    private final InetAddress address;
    private final String msg;

    ClientData(InetAddress address, String msg) {
        this.address = address;
        this.msg = msg;
    }

    public InetAddress getAddress() {
        return address;
    }

    public String getMsg() {
        return msg;
    }
}

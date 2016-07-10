function [bins] = fft_bins(data)
  n = rows(data);
  if n < 4
    bins = bin_matrix(n) * data;
  else
    eo = even_odd(n)*data;
    evenOut = fft_bins(eo(1:(n/2)));
    oddOut = fft_bins(eo((n/2+1):n));
    bins = combiner_matrix(n)*[evenOut; oddOut];
  end
end
